package contracts

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/pkg/rerror"
)

type Response[T any] struct {
	Status string          `json:"status"`
	Data   T               `json:"data"`
	Errors []ResponseError `json:"errors"`
}

type ResponseError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`

	InnerError error `json:"-"`
}

func (r ResponseError) Error() string {
	return r.Message
}

func StatusForKind(kind rerror.Kind) int {
	switch kind {
	case rerror.Permission:
		return http.StatusUnauthorized
	case rerror.Validation:
		return http.StatusBadRequest
	case rerror.Forbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func ResponseErrorFromError(error rerror.Error) ResponseError {
	if error.Kind() == rerror.Internal {
		return ResponseError{
			StatusCode: StatusForKind(error.Kind()),
			Message:    "Internal server error. Please try again later",
			InnerError: &error,
		}
	}
	return ResponseError{
		StatusCode: StatusForKind(error.Kind()),
		Message:    error.Error(),
	}
}
