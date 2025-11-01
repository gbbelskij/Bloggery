package custom_errors

import "errors"

var (
	ErrUserDoesNotExist = errors.New("user does not exist")
	ErrInvalidPassword  = errors.New("invalid password")
)
