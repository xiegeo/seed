package seed

import (
	"github.com/xiegeo/seed/dictionary"
)

// *Domain is a DomainGetter
var _ DomainGetter = &Domain{}

type ThingGetter interface {
	GetName() CodeName
	GetLabel() I18nGetter[string]
	GetDescription() I18nGetter[string]
}

func NewThing(from ThingGetter) Thing {
	return Thing{
		Name:        from.GetName(),
		Label:       NewI18n(from.GetLabel()),
		Description: NewI18n(from.GetDescription()),
	}
}

func (t Thing) GetName() CodeName {
	return t.Name
}

func (t Thing) GetLabel() I18nGetter[string] {
	return t.Label
}

func (t Thing) GetDescription() I18nGetter[string] {
	return t.Description
}

func (t Thing) GetThing() ThingGetter {
	return t
}

// DomainGetter describes a read only interface to a domain, upto the level of Field.
// As a compormized between interface complexity and extenablity of domain meta data.
type DomainGetter interface {
	ThingGetter
	GetObjects() dictionary.Getter[CodeName, ObjectGetter]
}

func (d *Domain) GetObjects() dictionary.Getter[CodeName, ObjectGetter] {
	return dictionary.MapValue[CodeName, *Object](d.Objects, func(v *Object) ObjectGetter {
		return v
	})
}

type ObjectGetter interface {
	ThingGetter
	FieldGroupGetter
}

type FieldGroupGetter interface {
	GetFields() dictionary.Getter[CodeName, *Field]
	GetIdentities() []Identity
	GetRanges() []Range
}

func (g *FieldGroup) GetFields() dictionary.Getter[CodeName, *Field] {
	return g.Fields
}

func (g *FieldGroup) GetIdentities() []Identity {
	return g.Identities
}

func (g *FieldGroup) GetRanges() []Range {
	return g.Ranges
}
