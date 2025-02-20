package entity

import "errors"

var (
	ErrAlreadyCreated = errors.New("user is already created")
	ErrIsNotExist     = errors.New("can't find such user")
)
