package seed

import (
	"math/big"
	"time"

	"github.com/shopspring/decimal"
)

type ObjectValue struct {
	Fields map[CodeName]FieldValue
}

type FieldValue struct {
	value any
}

type ReferenceValue struct {
	object CodeName
	id     int64
	ObjectValue
}

type SetFieldOption struct{}

/*
func (ob *ObjectValue) SetField(field *Field, value any, opt *SetFieldOption) error {
	// todo: use opt to optionally verify or transform value
	ob.Fields[field.Name] = FieldValue{value: value}
	return nil
}*/

type FieldValueType interface {
	string | *string | I18n[string] | // String FieldType
		[]byte | // Binary
		bool | *bool | // Boolean
		time.Time | *time.Time | // TimeStamp
		int64 | *int64 | *big.Int | // Integer
		float64 | *float64 | decimal.Decimal | *decimal.Decimal | // Real
		ReferenceValue | // Reference
		[]FieldValue | // List
		ObjectValue // Combination
}

func SetField[T FieldValueType](ob *ObjectValue, field *Field, value T, opt *SetFieldOption) error {
	// todo: use opt to optionally verify or transform value
	ob.Fields[field.Name] = FieldValue{value: value}
	return nil
}

func GetField[T FieldValueType](ob *ObjectValue, field *Field) (T, error) {
	var vt T //nolint:varnamelen
	v, ok := ob.Fields[field.Name]
	if !ok {
		return vt, NewFieldNotFoundError(field.Name)
	}
	vt, ok = v.value.(T)
	if ok { // fast pass
		return vt, nil
	}
	// todo: support type conversion
	return vt, NewTargetValueTypeNotSupportedError(field.Name, v, vt)
}
