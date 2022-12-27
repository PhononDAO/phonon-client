package backend

import "errors"

var (
	ErrCardUninitialized = errors.New("card uninitialized")
	ErrPhononTableFull   = errors.New("phonon table full")
	ErrKeyIndexInvalid   = errors.New("key index out of valid range")
	ErrOutOfMemory       = errors.New("card out of memory")
	ErrPINNotEntered     = errors.New("valid PIN required")
	ErrUnknown           = errors.New("unknown error")
)
