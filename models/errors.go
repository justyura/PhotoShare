package models

import "errors"

var (
    ErrNotFound = errors.New("models: resouce could not be found")
    ErrEmailTaken = errors.New("models: email address is already in use")
)
