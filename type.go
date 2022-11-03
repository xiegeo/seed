package seed

import (
	"math/big"
	"time"
)

// Thing is a base type for anything that can be identified.
type Thing struct {
	Name        CodeName // name is the long term api name of the thing, name is locally unique.
	Label       I18n[string]
	Discription I18n[string]
}

type CodeName string

func (c CodeName) String() string {
	return string(c)
}

// Domain holds a collection of objects, equivalent to a SQL database.
// Only one Domain is needed for most use cases.
type Domain struct {
	Thing
	Objects []*Object
}

type Object struct {
	Thing
	Fields     []*Field
	SubObjects []CodeName // Name of objects that supports all the fields. Mirrors subclass/implements by in OO.
}

type Field struct {
	Thing
	FieldType
	FieldTypeSetting
	IsI18n     bool // if true, different values for different locals is possible. Only String and Binary need to be supported.
	IsNullable bool // if true, difference between null and zero values are significate.
}

type FieldType int8

const (
	String FieldType = iota + 1
	Binary
	Boolean
	TimeStamp
	Integer
	Real
	Reference
	List
	Combination
)

/*
	 FieldTypeSetting is any of
		String      *StringSetting
		Binary      *BinarySetting
		Boolean     *BooleanSetting
	    TimeStamp   *TimeStampSetting
		Integer     *IntegerSetting
		Real        *RealSetting
		Reference   *ReferenceSetting
		List        *ListSetting
		Combination *CombinationSetting
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
	Standered   RealStandered
	MinMantissa *big.Int
	MaxMantissa *big.Int
	Base        *uint8
	MinExponent *int64
	MaxExponent *int64
	Unit        *Unit
}

type RealStandered int8

const (
	CustomReal RealStandered = iota
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
	Fields []*Field
}
