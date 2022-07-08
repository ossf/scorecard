package depdiff

import (
	"errors"
)

var (
	// ErrInvalidDepDiffFormat indicates the specified dependency diff output format is not valid.
	ErrInvalidDepDiffFormat = errors.New("invalid depdiff format")

	// ErrInvalidDepDiffFormat indicates the specified dependency diff output format is not valid.
	ErrMarshalDepDiffToJSON = errors.New("error marshal results to JSON")
)
