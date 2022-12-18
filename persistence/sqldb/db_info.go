package sqldb

import (
	"github.com/xiegeo/must"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/dictionary"
	"github.com/xiegeo/seed/seederrors"
)

type domainInfo struct {
	seed.Thing
	objectMap *dictionary.SelfKeyed[seed.CodeName, *objectInfo]
}

func (db *DB) domainInfoFromDomain(d seed.DomainGetter) (*domainInfo, error) {
	objectMap, err := seed.NewObjects[*objectInfo]()
	if err != nil {
		return nil, err
	}
	builder := db.newObjectInfoBuilder(d)
	err = d.GetObjects().RangeLogical(func(cn seed.CodeName, ob seed.ObjectGetter) (err error) {
		// the exact ordering does not mater, range deterministically to ease debugging.
		obInfo, err := builder.objectInfoFromObject(ob)
		if err != nil {
			return err
		}
		return objectMap.Add(cn, obInfo)
	})
	if err != nil {
		return nil, err
	}
	return &domainInfo{
		Thing:     seed.NewThing(d),
		objectMap: objectMap,
	}, nil
}

func (d *domainInfo) GetObjects() dictionary.Getter[seed.CodeName, seed.ObjectGetter] {
	return dictionary.MapValue[seed.CodeName, *objectInfo](d.objectMap, func(ob *objectInfo) seed.ObjectGetter {
		return ob
	})
}

type objectInfo struct {
	seed.Thing
	fields       *dictionary.SelfKeyed[seed.CodeName, *fieldInfo]
	mainTable    *Table
	helperTables map[string]*Table

	identities []seed.Identity
	ranges     []seed.Range
}

type objectInfoBuilder struct {
	db     *DB
	domain seed.DomainGetter
	object seed.ObjectGetter
}

type fieldInfoBuilder struct {
	*objectInfoBuilder
	parent *Table
}

func (db *DB) newObjectInfoBuilder(d seed.DomainGetter) *objectInfoBuilder {
	return &objectInfoBuilder{
		db:     db,
		domain: d,
	}
}

func (builder *objectInfoBuilder) objectInfoFromObject(ob seed.ObjectGetter) (*objectInfo, error) {
	builder.object = ob
	fields := seed.NewFields0[*fieldInfo]()
	table := builder.initTable()
	err := builder.setTableConstraints(table)
	if err != nil {
		return nil, err
	}
	childBuilder := &fieldInfoBuilder{
		objectInfoBuilder: builder,
		parent:            table,
	}
	helpers := make(map[string]*Table)
	revertColumnName := ExternalColumnName("").Revert
	err = ob.GetFields().RangeLogical(func(cn seed.CodeName, f *seed.Field) error {
		info, err := childBuilder.generateFieldInfo(f) //nolint:govet // shadow: declaration of "err"
		if err != nil {
			return err
		}
		err = fields.Add(f.GetName(), info)
		if err != nil {
			return seederrors.NewSystemError("error adding field that was already checked: %w", err)
		}
		for _, col := range info.cols {
			_, present := table.Columns.Set(revertColumnName(col.Name), col)
			if present {
				return seederrors.NewSystemError(`column with name="%s" inserted again, this should never happen`, col.Name)
			}
		}
		table.Constraint.Checks = append(table.Constraint.Checks, info.checks...)
		for _, helper := range info.tables {
			_, present := helpers[helper.TableName()]
			if present {
				return seederrors.NewSystemError(`table with name="%s" inserted again, this should never happen`, helper.Name)
			}
			helpers[helper.TableName()] = helper
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	table.Constraint.Checks = append(table.Constraint.Checks, getRangeChecks(ob)...)
	return &objectInfo{
		Thing:        seed.NewThing(ob),
		fields:       fields,
		mainTable:    table,
		helperTables: helpers,
	}, nil
}

func CheckPrimaryKey(f *seed.Field) error {
	switch f.FieldTypeSetting.(type) {
	default:
		return nil
	case seed.ListSetting:
		// system error for now
		return seederrors.NewSystemError("field %s with setting type %T can not be used as primary key", f.Name, f.FieldTypeSetting)
	}
}

func (builder *objectInfoBuilder) setTableConstraints(table *Table) error {
	table.Option = table.Option.Add(builder.db.option.TableOption)
	pkIndex, pkColumn, err := builder.db.option.getPrimaryKeys(builder.object)
	if err != nil {
		return err
	}
	if pkColumn == "" {
		table.Option = table.Option.Add(builder.db.option.TableOptionNoAutoID)
		if pkIndex < 0 || pkIndex >= len(builder.object.GetIdentities()) {
			return seederrors.NewSystemError(
				"index or column required for PrimaryKeys, got index %d, expected [0,%d)", pkIndex, len(builder.object.GetIdentities()))
		}
	} else {
		_, present := table.Columns.Set(systemColumnID, Column{
			Name: systemColumnID,
			Type: pkColumn,
			Constraint: ColumnConstraint{
				PrimaryKey: true,
			},
		})
		if present {
			return seederrors.NewSystemError(`column with name="%s" inserted again, this should never happen`, systemColumnID)
		}
		if pkIndex >= 0 {
			return seederrors.NewSystemError("both index=%d and column given for PrimaryKeys", pkIndex)
		}
	}
	idChecks := getIdentityChecks(builder.object)
	if pkIndex >= 0 {
		table.Constraint.PrimaryKeys, idChecks = cutOut(pkIndex, idChecks)
	}
	table.Constraint.Uniques = append(table.Constraint.Uniques, idChecks...)
	return nil
}

func (ob *objectInfo) GetFields() dictionary.Getter[seed.CodeName, *seed.Field] {
	return dictionary.MapValue[seed.CodeName, *fieldInfo](ob.fields, func(f *fieldInfo) *seed.Field {
		return &f.Field
	})
}

func (ob *objectInfo) GetIdentities() []seed.Identity {
	return ob.identities
}

func (ob *objectInfo) GetRanges() []seed.Range {
	return ob.ranges
}

func getRangeChecks(ob seed.ObjectGetter) []Expression {
	var exps []Expression
	must.NoError(seed.RangeRanges(ob, func(r seed.Range) error {
		a := "<"
		if r.IncludeEndValue {
			a = "<="
		}
		exps = append(exps,
			Expression{
				Type: BinaryExpression,
				A:    a,
				Expressions: []Expression{
					ValueLiteral(string(r.Start)),
					ValueLiteral(string(r.End)),
				},
			},
		)
		return nil
	}))
	return exps
}

func cutOut[V any](i int, v []V) (V, []V) {
	return v[i], append(v[:i], v[i+1:]...)
}

func getIdentityChecks(ob seed.ObjectGetter) [][]string {
	uniques := make([][]string, 0, len(ob.GetIdentities()))
	for _, ids := range ob.GetIdentities() {
		keys := make([]string, 0, len(ids.Fields)+len(ids.Ranges))
		for _, id := range ids.Fields {
			keys = append(keys, string(id))
		}
		for _, r := range ids.Ranges {
			keys = append(keys, string(r.Start))
		}
		must.True(len(keys) > 0, "an identity must have fields listed")
		uniques = append(uniques, keys)
	}
	return uniques
}

type fieldInfo struct {
	seed.Field
	fieldDefinition
	encoder func(any) ([]any, error)
	decoder func([]any) (any, error)
}

func (f *fieldInfo) Encoder() func(any) ([]any, error) {
	if f.encoder == nil {
		return func(value any) ([]any, error) {
			return []any{value}, nil
		}
	}
	return f.encoder
}

func (f *fieldInfo) Decoder() func([]any) (any, error) {
	if f.encoder == nil {
		return func(cols []any) (any, error) {
			if len(cols) != 1 {
				return nil, seederrors.NewSystemError("decoder is not defined on a fieldInfo with %d columns", len(cols))
			}
			return cols[0], nil
		}
	}
	return f.decoder
}

func (f *fieldInfo) WarpEncoder(fc func(any) (any, error)) {
	next := f.Encoder()
	f.encoder = func(value any) ([]any, error) {
		value, err := fc(value)
		if err != nil {
			return nil, err
		}
		return next(value)
	}
}

func (f *fieldInfo) WarpDecoder(fc func(any) (any, error)) {
	before := f.Decoder()
	f.decoder = func(cols []any) (any, error) {
		value, err := before(cols)
		if err != nil {
			return nil, err
		}
		return fc(value)
	}
}
