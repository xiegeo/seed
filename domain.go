package seed

import "github.com/xiegeo/seed/dictionary"

// CodeName marks a string as expected to follow seed naming rules.
// See ./dictionary for details.
type CodeName string

// Thing is a base type for anything that can be identified.
type Thing struct {
	Name        CodeName     // name is the long term api name of the thing, name is locally unique.
	Label       I18n[string] // used for displaying input label or column header.
	Description I18n[string] // used for displaying addition information.
}

// Domain holds a collection of objects, equivalent to all create table statements in a SQL database.
// Only one Domain is needed for most use cases.
//
// Domain is expected to be build once and never modified after first use. Data migration to support
// changing domain will not be done through direct modifications to domain.
type Domain struct {
	Thing
	Objects *dictionary.SelfKeyed[CodeName, *Object]
}

func NewDomain(thing Thing, objs ...*Object) (*Domain, error) {
	dict, err := NewObjects(objs...)
	if err != nil {
		return nil, err
	}
	return &Domain{
		Thing:   thing,
		Objects: dict,
	}, nil
}

func NewObjects[T ObjectGetter](objs ...T) (*dictionary.SelfKeyed[CodeName, T], error) {
	dict := NewObjects0[T]()
	err := dict.AddValue(objs...)
	if err != nil {
		return nil, err
	}
	return dict, nil
}

// NewObjects0 is the zero argument version of NewObjects, it also does not error
func NewObjects0[T ThingGetter]() *dictionary.SelfKeyed[CodeName, T] {
	return dictionary.NewSelfKeyed(
		dictionary.NewObject[CodeName, T](),
		func(ob T) CodeName {
			return ob.GetName()
		},
	)
}
