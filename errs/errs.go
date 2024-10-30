package errs

import "errors"

var (
	ErrRatingOutOfBounds = errors.New("error! Rating out of bounds")
)
