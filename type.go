package seed

import (
	"fmt"
	"math/big"
	"time"
)

type CodeName string

// Thing is a base type for anything that can be identified.
type Thing struct {
	Name        CodeName // name is the long term api name of the thing, name is locally unique.
	Label       I18n[string]
	Discription I18n[string]
}

// Domain holds a collection of objects, equivalent to a SQL database.
// Only one Domain is needed for most use cases.
type Domain struct {
	Thing
	Objects []Object
}

type Object struct {
	Thing
	Fields []Field

	// some form of class grouping can be useful, but not sure how to accomplish it yet.
	// SubObjects []CodeName // Name of objects that supports all the fields. Mirrors subclass/implements by in OO.
}

type Field struct {
	Thing
	FieldType
	FieldTypeSetting
	IsI18n   bool // if true, different values for different locals is possible. Only String and Binary need to be supported.
	Nullable bool // if true, difference between null and zero values are significate.
}

type FieldType int8

const (
	FieldTypeUnset FieldType = iota
	String
	Binary
	Boolean
	TimeStamp
	Integer
	Real
	Reference
	List
	Combination
	FieldTypeMax = Combination
)

var _fieldTypeStringer = []string{"FieldTypeUnset", "String", "Binary", "Boolean", "TimeStamp", "Integer", "Real", "Reference", "List", "Combination"}

func (f FieldType) String() string {
	if f < 0 || f > FieldTypeMax {
		return fmt.Sprintf("FieldType(%d) out of range[%d,%d]", f, FieldTypeUnset, FieldTypeMax)
	}
	return _fieldTypeStringer[f]
}

func (f FieldType) Valid() bool {
	return f <= FieldTypeUnset || f > FieldTypeMax
}

/*
FieldTypeSetting is any of:

		String      StringSetting
		Binary      BinarySetting
		Boolean     BooleanSetting
	    TimeStamp   TimeStampSetting
		Integer     IntegerSetting
		Real        RealSetting
		Reference   ReferenceSetting
		List        ListSetting
		Combination CombinationSetting
*/
type FieldTypeSetting any

type StringSetting struct {
	MaxCodePoints int64
	IsSingleLine  bool
}

type BinarySetting struct {
	MaxBytes int64
}

type BooleanSetting struct{}

type TimeStampSetting struct {
	Min      time.Time
	Max      time.Time
	Accuracy time.Duration // >= 24h: date only; >= 1s datetime
}

type IntegerSetting struct {
	Min  *big.Int
	Max  *big.Int
	Unit *Unit
}

type Unit struct {
	Thing
	Symble string
}

type RealSetting struct {
	Standard    RealStandard
	MinMantissa *big.Int
	MaxMantissa *big.Int
	Base        *uint8
	MinExponent *int64
	MaxExponent *int64
	MinFloat    *float64
	MaxFloat    *float64
	Unit        *Unit
}

type RealStandard int8

const (
	CustomReal RealStandard = iota
	Float32
	Float64
	Decimal32
	Decimal64
)

type ReferenceSetting struct {
	ObjectsAllowed  []CodeName // list of objects that can be referenced.
	IsWeakReference bool       // If false, the target must exist.
}

// ListSetting describes a collection of the same type:
// IsOrdered | IsUnique | collection type
// false | false | counted set
// false | true | set
// true | any | list
type ListSetting struct {
	MaxLength int64
	IsOrdered bool // if true, the items ordering is preserved.
	IsUnique  bool // if true, repeated items should be ignored.

	ItemType        FieldType
	ItemTypeSetting FieldTypeSetting

	// The following options might have no use.
	// IsI18n          bool // if true, different values per item for different locals is possible.
	// IsNullable      bool // if true, list include null elements.
}

type CombinationSetting struct {
	Fields []Field
}
