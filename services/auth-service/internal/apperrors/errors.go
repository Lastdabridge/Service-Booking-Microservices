package apperrors

import (
	"errors"
	"net/http"
)

type Err struct {
	StatusCode int
	Msg        string
}

var (
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrNameHasSpecialChars = errors.New("name must not contain special characters")
	ErrPasswordHashFailed  = errors.New("failed to hash password")

	ErrInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrAccountNotFound        = errors.New("account not found")
)

var errsMap = map[error]Err{
	ErrAccountNotFound: {
		StatusCode: http.StatusNotFound, //404
		Msg:        ErrAccountNotFound.Error(),
	},
	ErrUserAlreadyExists: {
		StatusCode: http.StatusConflict, // 409
		Msg:        ErrUserAlreadyExists.Error(),
	},
	ErrNameHasSpecialChars: {
		StatusCode: http.StatusBadRequest, // 400
		Msg:        ErrNameHasSpecialChars.Error(),
	},
	ErrPasswordHashFailed: {
		StatusCode: http.StatusInternalServerError, // 500
		Msg:        ErrPasswordHashFailed.Error(),
	},
	ErrInvalidLoginOrPassword: {
		StatusCode: http.StatusUnauthorized, // 401
		Msg:        ErrInvalidLoginOrPassword.Error(),
	},
}

func Get(err error) Err {
	if e, ok := errsMap[err]; ok {
		return e
	}

	return Err{
		StatusCode: http.StatusInternalServerError,
		Msg:        "internal server error",
	}
}
