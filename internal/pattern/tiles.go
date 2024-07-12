package pattern

import "errors"

var (
	ErrInvalidHeader     = errors.New("invalid header")
	ErrMissingTerminator = errors.New("missing terminator")
	ErrOverflow          = errors.New("overflow")
)
