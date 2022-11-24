package errors

import (
	"errors"
)

// Error implements build-in error interface.
type Error struct {
	Code    Code
	Details string
}

func (e Error) Error() string {
	msg := e.Code.String()
	if len(e.Details) > 0 {
		msg += ": " + e.Details
	}
	return msg
}

// As finds the first error in err's chain that matches Error type and returns it.
func As(err error) (Error, bool) {
	var e Error
	if errors.As(err, &e) {
		return e, true
	}
	pe := new(Error)
	if errors.As(err, &pe) {
		return *pe, true
	}
	return Error{}, false
}
