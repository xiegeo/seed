package seederrors

import (
	"github.com/cockroachdb/errors/errutil"
	"github.com/cockroachdb/errors/secondary"
)

// WithMessagef annotates err with the format specifier.
// If err is nil, WithMessagef returns nil.
func WithMessagef(err error, format string, args ...interface{}) error {
	return errutil.WithMessagef(err, format, args...)
}



// CombineErrors returns err, or, if err is nil, otherErr.
// if err is non-nil, otherErr is attached as secondary error.
func CombineErrors(err error, otherErr error) error {
	return secondary.CombineErrors(err, otherErr)
}
