package util

import (
	"errors"
	"net/http"
)

type UserFacingError interface {
	error
	HttpRet() int
}
type userFacingError struct {
	error
	httpRet int
}

func (u userFacingError) HttpRet() int {
	return u.httpRet
}

func ErrorFromHttpError(statusCode int, errorString string) UserFacingError {
	return userFacingError{
		error:   errors.New(errorString),
		httpRet: statusCode,
	}
}

func BadInputError(err error) UserFacingError {
	return userFacingError{
		error:   err,
		httpRet: http.StatusBadRequest,
	}
}

func ForbiddenError(err error) UserFacingError {
	return userFacingError{
		error:   err,
		httpRet: http.StatusForbidden,
	}
}

func DuplicateInputError(err error) UserFacingError {
	return userFacingError{
		error:   err,
		httpRet: http.StatusConflict,
	}
}

func NotFoundError(err error) UserFacingError {
	return userFacingError{
		error:   err,
		httpRet: http.StatusNotFound,
	}
}
