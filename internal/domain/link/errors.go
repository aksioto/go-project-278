package link

import "errors"

var (
	ErrNotFound           = errors.New("link not found")
	ErrShortNameTaken     = errors.New("short name already exists")
	ErrVisitNotFound      = errors.New("visit not found")
	ErrInvalidID          = errors.New("invalid id")
	ErrInvalidOriginalURL = errors.New("cannot shorten own short URLs")
)
