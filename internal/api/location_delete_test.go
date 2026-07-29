package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_DeleteLocation(t *testing.T) {
	t.Run("204_when_the_location_belongs_to_the_users_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		location := fixtures.NewLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))

		w := env.do(http.MethodDelete, "/v1/locations/"+location.ID.String(), token, nil)
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("403_when_the_location_belongs_to_another_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		otherLocation := fixtures.NewLocation(t, env.cfg, env.bz)

		w := env.do(http.MethodDelete, "/v1/locations/"+otherLocation.ID.String(), token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodDelete, fmt.Sprintf("/v1/locations/%s", ksuid.New()), "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
