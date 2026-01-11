package link

import "errors"

var (
	ErrNotFound       = errors.New("link not found")
	ErrShortNameTaken = errors.New("short name already exists")
)
