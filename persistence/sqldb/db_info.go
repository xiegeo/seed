package sqldb

import (
	"context"

	"github.com/xiegeo/must"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/dictionary"
	"github.com/xiegeo/seed/seederrors"
)

type domainInfo struct {
	seed.Thing
	objectMap *dictionary.SelfKeyed[seed.CodeName, *objectInfo]
}

func (db *DB) domainInfoFromDomain(ctx context.Context, d seed.DomainGetter) (*domainInfo, error) {
	objectMap, err := seed.NewObjects[*objectInfo]()
	if err != nil {
		return nil, err
	}
	err = d.GetObjects().RangeLogical(func(cn seed.CodeName, ob seed.ObjectGetter) (err error) {
		// the exact ordering does not mater, range deterministically to ease debugging.
		obInfo, err := db.objectInfoFromObject(ctx, d, ob)
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
	mainTable    Table
	helperTables map[string]Table

	identities []seed.Identity
	ranges     []seed.Range
}

func (db *DB) objectInfoFromObject(ctx context.Context, d seed.DomainGetter, ob seed.ObjectGetter) (*objectInfo, error) {
	tableName, err := db.generateTableName(ctx, d, ob)
	if err != nil {
		return nil, err
	}
	fields := seed.NewFields0[*fieldInfo]()

	table := InitTable(tableName)
	table.Option = db.option.TableOption
	helpers := make(map[string]Table)
	err = ob.GetFields().RangeLogical(func(cn seed.CodeName, f *seed.Field) error {
		info, err := db.generateFieldInfo(f) //nolint:govet // shadow: declaration of "err"
		if err != nil {
			return err
		}
		err = fields.Add(f.GetName(), info)
		if err != nil {
			return seederrors.NewSystemError("error adding field that was already checked: %w", err)
		}
		for _, col := range info.cols {
			_, present := table.Columns.Set(col.Name, col)
			if present {
				return seederrors.NewSystemError(`column with name="%s" inserted again, this should never happen`, col.Name)
			}
		}
		table.Constraint.Checks = append(table.Constraint.Checks, info.checks...)
		for _, helper := range info.tables {
			_, present := helpers[helper.Name]
			if present {
				return seederrors.NewSystemError(`table with name="%s" inserted again, this should never happen`, helper.Name)
			}
			helpers[helper.Name] = helper
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
