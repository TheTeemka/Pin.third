package models

import "third/merrors"

var (
	ErrNotFound   = merrors.New("models: resource could not be found")
	ErrEmailTaken = merrors.New("models: email address is already in use")
)
