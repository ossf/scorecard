package depdiff

import (
	"errors"
)

var (
	// ErrInvalidDepDiffFormat indicates the specified dependency diff output format is not valid.
	ErrInvalidDepDiffFormat = errors.New("invalid depdiff format")

	// ErrDepDiffFormatNotSupported indicates the specified dependency diff output format is not supported.
	ErrDepDiffFormatNotSupported = errors.New("depdiff format not supported")

	// ErrInvalidDepDiffFormat indicates the specified dependency diff output format is not valid.
	ErrMarshalDepDiffToJSON = errors.New("error marshal results to JSON")
)
