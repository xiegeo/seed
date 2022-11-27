package seed

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/xiegeo/seed/seederrors"
)

type CodeName string

// Thing is a base type for anything that can be identified.
type Thing struct {
	Name        CodeName     // name is the long term api name of the thing, name is locally unique.
	Label       I18n[string] // used for input label or column header
	Description I18n[string] // addition information
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
	FieldTypeSetting
	FieldType
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
	if f <= FieldTypeUnset || f > FieldTypeMax {
		return invalid()
	}
	return true
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

func GetFieldTypeSetting[T FieldTypeSetting](f *Field) (T, error) {
	v, ok := f.FieldTypeSetting.(T)
	if !ok {
		return v, seederrors.NewSystemError("can not get a %T from field %s with type %s and setting of %T", v, f.Name, f.FieldType, f.FieldTypeSetting)
	}
	return v, nil
}

func peek[T FieldTypeSetting](s FieldTypeSetting) T {
	v, _ := s.(T)
	return v
}

func FieldTypeSettingCover(s, s2 FieldTypeSetting) bool {
	switch vt := s.(type) {
	case StringSetting:
		return vt.Covers(peek[StringSetting](s2))
	case BinarySetting:
		return vt.Covers(peek[BinarySetting](s2))
	case BooleanSetting:
		return vt.Covers(peek[BooleanSetting](s2))
	case TimeStampSetting:
		return vt.Covers(peek[TimeStampSetting](s2))
	case IntegerSetting:
		return vt.Covers(peek[IntegerSetting](s2))
	case RealSetting:
		return vt.Covers(peek[RealSetting](s2))
	case ReferenceSetting:
		return vt.Covers(peek[ReferenceSetting](s2))
	case ListSetting:
		return vt.Covers(peek[ListSetting](s2))
	case CombinationSetting:
		return false
		// return vt.Covers(peek[CombinationSetting](s2)) // future
	}
	return invalid()
}

type StringSetting struct {
	MinCodePoints int64
	MaxCodePoints int64
	IsSingleLine  bool
}

// Covers returns true if s can support all values in s2
func (s StringSetting) Covers(s2 StringSetting) bool {
	switch {
	case
		s.MinCodePoints > s2.MinCodePoints,
		s.MaxCodePoints < s2.MaxCodePoints,
		s.IsSingleLine && !s2.IsSingleLine:
		return false
	}
	return true
}

type BinarySetting struct {
	MinBytes int64
	MaxBytes int64
}

// Covers returns true if s can support all values in s2
func (s BinarySetting) Covers(s2 BinarySetting) bool {
	switch {
	case
		s.MinBytes > s2.MinBytes,
		s.MaxBytes < s2.MaxBytes:
		return false
	}
	return true
}

type BooleanSetting struct{}

// Covers returns true if s can support all values in s2
func (s BooleanSetting) Covers(s2 BooleanSetting) bool {
	return true
}

type TimeStampSetting struct {
	Min                time.Time
	Max                time.Time
	Scale              time.Duration // >= 24h: date only; >= 1s: datetime; < 1s: support fraction seconds
	WithTimeZoneOffset bool          // if false always UTC
}

// Covers returns true if s can support all values in s2
func (s TimeStampSetting) Covers(s2 TimeStampSetting) bool {
	switch {
	case
		s.Min.After(s2.Min),
		s.Max.Before(s2.Max),
		s.Scale > s2.Scale,
		!s.WithTimeZoneOffset && s2.WithTimeZoneOffset:
		return false
	}
	return true
}

type IntegerSetting struct {
	Min  *big.Int
	Max  *big.Int
	Unit *Unit
}

// Covers returns true if s can support all values in s2
func (s IntegerSetting) Covers(s2 IntegerSetting) bool {
	switch {
	case
		s.Min.Cmp(s2.Min) == 1,
		s.Max.Cmp(s2.Max) == -1:
		return false
	}
	return s.Unit.Covers(s2.Unit)
}

type Unit struct {
	Thing
	Symble string // a display symbol, such as: %, °C
}

// Covers returns true if s can support all values in s2.
//
// For unit comparisons, no value ranges are checked, So *Unit.Covers has a different meaning.
// For most use cases, receiver s should be nil, since it describes the ability to handle a set
// of values without care to the unit. But in case it does care about the unit, we protect against
// mixing units by returning false if s2 is different from s.
func (s *Unit) Covers(s2 *Unit) bool {
	if s == nil || s == s2 {
		return true
	}
	if s2 == nil {
		return false
	}
	return s.Name == s2.Name && s.Symble == s2.Symble
}

type RealSetting struct {
	Standard    RealStandard
	Base        uint8 // base 2 and 10 are most common, others are unlikely to be supported.
	MinMantissa *big.Int
	MaxMantissa *big.Int
	MinExponent *int64 // pointer to diff zero vs not set
	MaxExponent *int64

	// Alternative settings for RealStandard of Float32 or Float64 to replace *Mantissa and *Exponent.
	// The full range is supported if not set.
	MinFloat *float64
	MaxFloat *float64

	Unit *Unit
}

func (s *RealSetting) Valid() bool {
	switch s.Standard {
	case CustomReal:
		switch {
		case
			s.MinMantissa == nil,
			s.MaxMantissa == nil,
			s.Base < 2,
			s.MinExponent == nil,
			s.MaxExponent == nil:
			return invalid()
		}
	case Float32, Float64:
		if s.MinFloat == nil {
			s.MinFloat = valuePointer(math.Inf(-1))
		}
		if s.MaxFloat == nil {
			s.MaxFloat = valuePointer(math.Inf(1))
		}
	}
	return true
}

// invalid is called before Valid returns false to help easily break on bad settings.
func invalid() bool {
	return false
}

// valuePointer returns a pointer for any input, be careful that input is passed by value.
func valuePointer[T any](v T) *T {
	return &v
}

// Covers returns true if s can support all values in s2.
func (s RealSetting) Covers(s2 RealSetting) bool {
	switch {
	case
		!s.Valid(), !s2.Valid(),
		!s.Unit.Covers(s2.Unit),
		!s.Standard.Covers(s2.Standard):
		return false
	}
	if s.Standard == CustomReal && s2.Standard == CustomReal {
		switch {
		case
			s.MinMantissa.Cmp(s2.MinMantissa) == 1,
			s.MaxMantissa.Cmp(s2.MaxMantissa) == -1,
			s.Base != s2.Base,
			*s.MinExponent > *s2.MinExponent,
			*s.MaxExponent < *s2.MaxExponent:
			return false
		}
		return true
	}
	// only the float32/64 cases are left
	switch {
	case
		*s.MinFloat > *s2.MinFloat,
		*s.MaxFloat < *s2.MaxFloat:
		return false
	}
	return true
}

type RealStandard int8

const (
	CustomReal RealStandard = iota
	Float32
	Float64
	// Decimal32
	// Decimal64
)

// Covers returns true if s can support all values in s2.
func (s RealStandard) Covers(s2 RealStandard) bool {
	switch {
	case
		s == s2,
		// s == CustomReal, // future: allow custom type to support standard real types
		// s2 == s. // future: allow standard real types to support custom types
		s == Float64 && s2 == Float32:
		// s == Decimal64 && s2 == Decimal32:
		return true
	}
	return false
}

type ReferenceSetting struct {
	Target CodeName // target object.
	// IsWeakReference bool     // If false, the target must exist. If true, the deletion of target sets this reference to null.
}

// Covers returns true if s can support all values in s2.
func (s ReferenceSetting) Covers(s2 ReferenceSetting) bool {
	return s.Target == "" || s.Target == s2.Target
}

// ListSetting describes a collection of the same type:
// IsOrdered | IsUnique | collection type
// false | false | counted set
// false | true | set
// true | any | list
type ListSetting struct {
	MinLength int64
	MaxLength int64
	IsOrdered bool // if true, the items ordering is preserved.
	IsUnique  bool // if true, repeated items should be ignored.

	ItemType        FieldType
	ItemTypeSetting FieldTypeSetting

	// The following options might have no use.
	// IsI18n          bool // if true, different values per item for different locals is possible.
	// IsNullable      bool // if true, list include null elements.
}

func (s ListSetting) Covers(s2 ListSetting) bool {
	switch {
	case
		s.MinLength < s2.MinLength,
		s.MaxLength < s2.MaxLength,
		!s.IsOrdered && s2.IsOrdered,
		s.IsUnique && !s2.IsUnique,
		s.ItemType != s2.ItemType,
		!FieldTypeSettingCover(s.ItemTypeSetting, s2.ItemTypeSetting):
		return false
	}
	return true
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
