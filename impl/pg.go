package impl

import "github.com/lib/pq"

// pg error code defines.
const (
	UniqueViolation    pq.ErrorCode = "23505"
	MaxValueOfSequence pq.ErrorCode = "2200H"
)

// IsPqError check error is a pq error.
func IsPqError(err error, code pq.ErrorCode) bool {
	if e, ok := err.(*pq.Error); ok {
		return e.Code == code
	}
	return false
}
