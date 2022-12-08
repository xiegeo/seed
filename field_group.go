package seed

import (
	"github.com/xiegeo/seed/dictionary"
)

// Object describes a business object.
type Object struct {
	Thing
	FieldGroup
}

// FieldGroup describe a collection of fields.
type FieldGroup struct {
	Fields     *dictionary.SelfKeyed[CodeName, *Field] // List each fields. The ordering of fields is not relevant for behavior but preserved for implementation details.
	Identities []Identity                              // required for object definitions and CombinationSetting on fields referred to in parent identity.
	Ranges     []Range                                 // if any ranges is left out of identities, it can be described here.
}

// NewFields build a self keyed dictionary for use in FieldGroup.Fields
func NewFields(fs ...*Field) (*dictionary.SelfKeyed[CodeName, *Field], error) {
	dict := dictionary.NewSelfKeyed(
		dictionary.NewField[CodeName, *Field](),
		func(f *Field) CodeName {
			return f.Name
		},
	)
	err := dict.AddValue(fs...)
	if err != nil {
		return nil, err
	}
	return dict, nil
}

func (g *FieldGroup) RangeRanges(f func(r Range) error) error {
	for _, id := range g.Identities {
		for _, r := range id.Ranges {
			err := f(r)
			if err != nil {
				return err
			}
		}
	}
	for _, r := range g.Ranges {
		err := f(r)
		if err != nil {
			return err
		}
	}
	return nil
}

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
