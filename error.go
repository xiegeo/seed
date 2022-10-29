package seed

import (
	"fmt"
)

type FieldNotFoundError struct {
	FieldName CodeName
}

func NewFieldNotFoundError(fieldName CodeName) FieldNotFoundError {
	return FieldNotFoundError{FieldName: fieldName}
}

func (e FieldNotFoundError) Error() string {
	return fmt.Sprintf(`field "%s" is not found`, e.FieldName)
}

type TargetValueTypeNotSupportedError struct {
	FieldName CodeName
	Value     any
	Target    any
}

func NewTargetValueTypeNotSupportedError(fieldName CodeName, value, target any) TargetValueTypeNotSupportedError {
	return TargetValueTypeNotSupportedError{
		FieldName: fieldName,
		Value:     value,
		Target:    target,
	}
}

func (e TargetValueTypeNotSupportedError) Error() string {
	return fmt.Sprintf(`field "%s" can not value convert from %T to %T`, e.FieldName, e.Value, e.Target)
}
