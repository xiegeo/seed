package sqldb

import (
	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

// ColumnFeatures describe purposed column types that support each field
// ordered by preference. If none of the listed column features match, a
// automatic fail back algorithm is used. For example: integer column for
// boolean field.
type ColumnFeatures [seed.FieldTypeMax][]ColumnFeature

type ColumnFeature struct {
	TypeName        ColumnType            // sql column type name
	AcceptArguments bool                  // if standard arguments should be provide on sql column type definition
	Implement       seed.FieldTypeSetting // the widest allowed range of the field, optional feature support
}

func (c *ColumnFeatures) Append(typeName ColumnType, acceptArgs bool, f *seed.Field) error {
	if !f.FieldType.Valid() {
		return seederrors.NewSystemError("field type %s is not valid", f.FieldType)
	}
	c[f.FieldType-1] = append(c[f.FieldType-1], ColumnFeature{
		TypeName:        typeName,
		AcceptArguments: acceptArgs,
		Implement:       f.FieldTypeSetting,
	})
	return nil
}

func (c *ColumnFeatures) MustAppend(typeName ColumnType, acceptArgs bool, f *seed.Field) {
	panicOnError(c.Append(typeName, acceptArgs, f))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func (c *ColumnFeatures) Match(f *seed.Field) (ColumnFeature, bool) {
	if !f.FieldType.Valid() {
		return ColumnFeature{}, false
	}
	typeList := c[f.FieldType-1]
	for _, columnFeature := range typeList {
		if seed.FieldTypeSettingCover(columnFeature.Implement, f.FieldTypeSetting) {
			return columnFeature, true
		}
	}
	return ColumnFeature{}, false
}

func (c ColumnFeature) fieldDefinition(f *seed.Field) (fieldDefinition, error) {
	col := Column{
		Name: string(f.Name),
		Type: c.TypeName,
		Constraint: ColumnConstraint{
			NotNull: !f.Nullable,
		},
	}
	if c.AcceptArguments {
		return fieldDefinition{}, seederrors.NewSystemError("arguments not implemented for SQL column %s", c.TypeName)
	}
	// future: add checks to add additional safety.
	return fieldDefinition{
		cols: []Column{col},
	}, nil
}
