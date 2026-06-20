package errors

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("forbidden")
	ErrValidation = errors.New("validation error")
	ErrConflict   = errors.New("conflict")
)

func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return 404
	case errors.Is(err, ErrForbidden):
		return 403
	case errors.Is(err, ErrValidation):
		return 400
	case errors.Is(err, ErrConflict):
		return 409
	default:
		return 500
	}
}
