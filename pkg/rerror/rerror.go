package rerror

import (
	"errors"
	"runtime"
)

type Kind uint8

const (
	Internal Kind = iota
	Permission
	Validation
	Forbidden
)

func (k Kind) String() string {
	switch k {
	case Permission:
		return "permission"
	case Validation:
		return "validation"
	case Forbidden:
		return "forbidden"
	default:
		return "internal"
	}
}

type Error struct {
	Err error

	kind  Kind
	stack []uintptr
}

func (e *Error) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *Error) Kind() Kind {
	return e.kind
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(err error) *Error {
	return build(err, Internal)
}

func NewMessage(message string, kind Kind) *Error {
	return build(errors.New(message), kind)
}

func Wrap(err error) error {
	if err == nil {
		return nil
	}
	var re *Error
	if errors.As(err, &re) {
		return err
	}
	return New(err)
}

func (e *Error) WithKind(kind Kind) *Error {
	return &Error{
		Err:   e.Err,
		kind:  kind,
		stack: e.stack,
	}
}

func build(err error, kind Kind) *Error {
	return &Error{
		Err:   err,
		kind:  kind,
		stack: callers(),
	}
}

func callers() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(4, pcs[:])
	return pcs[:n]
}
