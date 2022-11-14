package seed

import (
	"fmt"
	"math/big"
	"time"
)

type CodeName string

// Thing is a base type for anything that can be identified.
type Thing struct {
	Name        CodeName     // name is the long term api name of the thing, name is locally unique.
	Label       I18n[string] // used for input label or column header
	Discription I18n[string] // addition information
}

// Domain holds a collection of objects, equivalent to a SQL database.
// Only one Domain is needed for most use cases.
type Domain struct {
	Thing
	Objects []Object
}

type Object struct {
	Thing
	FieldProperties

	// some form of class grouping can be useful, but not sure how to accomplish it yet.
	// SubObjects []CodeName // Name of objects that supports all the fields. Mirrors subclass/implements by in OO.
}

// FieldProperties describe a collection of fields.
type FieldProperties struct {
	Fields     []Field    // List each fields. The ordering of fields is not relevant for behavior.
	Identities []Identity // required for object definitions and CombinationSetting on fields referred to in identity.
	Ranges     []Range    // if any ranges is left out of identities, it can be described here.
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
	Min   time.Time
	Max   time.Time
	Scale time.Duration // >= 24h: date only; >= 1s datetime
}

type IntegerSetting struct {
	Min  *big.Int
	Max  *big.Int
	Unit *Unit
}

type Unit struct {
	Thing
	Symble string // a display symbol, such as: %, °C
}

type RealSetting struct {
	Standard    RealStandard
	MinMantissa *big.Int
	MaxMantissa *big.Int
	Base        *uint8 // base 2 and 10 are most common, others are unlikely to be supported.
	MinExponent *int64
	MaxExponent *int64
	MinFloat    *float64 // alternative settings for RealStandard of Float32 or Float64 to replace *Mantissa and *Exponent
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

type CombinationSetting FieldProperties // Reuse FieldSettings

// Identity is used to mark a subset of fields in an object or a combination field as capable of
// identifying an single instance (a license plate number can ID a car), or
// uniqueness is required for correct modeling of state (two philosophers can not use the same
// fork at the same time).
type Identity struct {
	Fields []CodeName
	Ranges []Range
}

// Range marks two fields by name as describing a range of values.
// Start and End must reference two comparable fields of the same type.
//
//   - If IncludeEndValue = false, then Start < End. (range must have none zero length)
//   - If IncludeEndValue = true, then Start <= End.
//
// Currently, the only for seen usage of Range is to support time ranges. Under this context,
// the end value of a time range can create ambiguities:
//
//   - When a room is booked from 1 to 2 o'clock, it can be booked from 2 onwards.
//     Here the end value is excluded from the range.
//   - When a person is busy from Monday to Friday, he is still busy on Friday.
//     Here the end value is included in the range.
//
// Although it's possible to only support one type of range on the backend and only convert to
// the users' expectation on display, this crates a horrible off-by-one trap that can only be
// fixed once on the display path. By adding an IncludeEndValue option, the most human friendly
// interpretation of end value is preserved thought-out.
type Range struct {
	Start           CodeName
	End             CodeName
	IncludeEndValue bool
}
