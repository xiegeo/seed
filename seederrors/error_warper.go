package seederrors

import "github.com/cockroachdb/errors/errutil"

// WithMessagef annotates err with the format specifier.
// If err is nil, WithMessagef returns nil.
func WithMessagef(err error, format string, args ...interface{}) error {
	return errutil.WithMessagef(err, format, args...)
}
