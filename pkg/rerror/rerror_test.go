package rerror_test

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/stretchr/testify/require"
)

func Test_Kind_String(t *testing.T) {
	t.Run("names_each_known_kind", func(t *testing.T) {
		require.Equal(t, "internal", Internal.String())
		require.Equal(t, "permission", Permission.String())
		require.Equal(t, "validation", Validation.String())
		require.Equal(t, "forbidden", Forbidden.String())
	})

	t.Run("falls_back_to_internal_for_an_unknown_kind", func(t *testing.T) {
		require.Equal(t, "internal", Kind(200).String())
	})
}

func Test_New(t *testing.T) {
	t.Run("builds_an_error_with_internal_kind_by_default", func(t *testing.T) {
		err := New(errors.New("email is required"))

		require.Equal(t, Internal, err.Kind())
		require.Equal(t, "email is required", err.Error())
	})

	t.Run("stores_the_error_as_the_underlying_error", func(t *testing.T) {
		innerErr := errors.New("boom")
		err := New(innerErr)

		require.Equal(t, innerErr, err.Err)
	})

	t.Run("creates_non_nil_error", func(t *testing.T) {
		err := New(errors.New("boom"))

		require.NotNil(t, err)
	})
}

func Test_Unwrap(t *testing.T) {
	sentinel := errors.New("not found")

	t.Run("lets_errors_is_traverse_to_the_wrapped_cause", func(t *testing.T) {
		err := New(sentinel)

		require.ErrorIs(t, err, sentinel)
	})

	t.Run("lets_errors_is_traverse_through_nested_wrapping", func(t *testing.T) {
		err := New(New(sentinel))

		require.ErrorIs(t, err, sentinel)
	})

	t.Run("outermost_rerror_kind_still_wins_for_errors_as", func(t *testing.T) {
		inner := New(sentinel).WithKind(Validation)
		outer := New(inner).WithKind(Forbidden)

		var re *Error
		require.ErrorAs(t, outer, &re)
		require.Equal(t, Forbidden, re.Kind())
	})
}

func Test_Wrap(t *testing.T) {
	t.Run("returns_nil_for_a_nil_error", func(t *testing.T) {
		require.NoError(t, Wrap(nil))
	})

	t.Run("wraps_a_foreign_error_as_internal", func(t *testing.T) {
		err := Wrap(errors.New("db down"))

		var re *Error
		require.ErrorAs(t, err, &re)
		require.Equal(t, Internal, re.Kind())
	})

	t.Run("preserves_the_kind_of_an_existing_rerror", func(t *testing.T) {
		err := Wrap(NewMessage("email is required", Validation))

		var re *Error
		require.ErrorAs(t, err, &re)
		require.Equal(t, Validation, re.Kind())
	})

	t.Run("preserves_the_kind_of_a_wrapped_rerror", func(t *testing.T) {
		err := Wrap(fmt.Errorf("outer: %w", NewMessage("denied", Forbidden)))

		var re *Error
		require.ErrorAs(t, err, &re)
		require.Equal(t, Forbidden, re.Kind())
	})
}

func Test_WithKind(t *testing.T) {
	t.Run("changes_the_kind_of_an_error", func(t *testing.T) {
		orig := New(errors.New("db down"))
		err := orig.WithKind(Permission)

		require.Equal(t, Permission, err.Kind())
		require.Equal(t, "db down", err.Error())
	})

	t.Run("preserves_the_underlying_error", func(t *testing.T) {
		origErr := errors.New("db down")
		orig := New(origErr)
		err := orig.WithKind(Permission)

		require.Equal(t, origErr, err.Err)
	})

	t.Run("creates_non_nil_error", func(t *testing.T) {
		orig := New(errors.New("boom"))
		err := orig.WithKind(Permission)

		require.NotNil(t, err)
	})

	t.Run("enables_method_chaining", func(t *testing.T) {
		err := New(errors.New("invalid input")).WithKind(Validation)

		require.Equal(t, Validation, err.Kind())
		require.Equal(t, "invalid input", err.Error())
	})

	t.Run("interoperates_with_fmt_errorf_wrapping", func(t *testing.T) {
		root := New(errors.New("no token")).WithKind(Permission)

		err := fmt.Errorf("outer: %w", root)

		var re *Error
		require.ErrorAs(t, err, &re)
		require.Equal(t, Permission, re.Kind())
	})
}
