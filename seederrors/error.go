// seederrors is mostly a warper around github.com/cockroachdb/errors
// to use a subset of the features.
package seederrors

import (
	"fmt"
)

type FieldNotFoundError struct {
	FieldName string
}

type stringer interface {
	string
	fmt.Stringer
}

func toString[S stringer](v S) string {
	switch vt := any(v).(type) {
	case string:
		return vt
	case fmt.Stringer:
		return vt.String()
	}
	panic("interface case handling exhausted")
}

func NewFieldNotFoundError[S stringer](fieldName S) FieldNotFoundError {
	return FieldNotFoundError{FieldName: toString(fieldName)}
}

func (e FieldNotFoundError) Error() string {
	return fmt.Sprintf(`field "%s" is not found`, e.FieldName)
}

type TargetValueTypeNotSupportedError struct {
	FieldName string
	Value     any
	Target    any
}

func NewTargetValueTypeNotSupportedError[S stringer](fieldName S, value, target any) TargetValueTypeNotSupportedError {
	return TargetValueTypeNotSupportedError{
		FieldName: toString(fieldName),
		Value:     value,
		Target:    target,
	}
}

func (e TargetValueTypeNotSupportedError) Error() string {
	return fmt.Sprintf(`field "%s" can not value convert from %T to %T`, e.FieldName, e.Value, e.Target)
}
