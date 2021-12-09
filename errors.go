package distlock

import "errors"

var (
	ErrAlreadyLocked = errors.New("already locked")
	ErrNotLocked     = errors.New("not locked")
)
