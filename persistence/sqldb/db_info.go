package sqldb

import (
	"context"

	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

type domainInfo struct {
	seed.Thing
	objectMap map[seed.CodeName]objectInfo
}

func (db *DB) domainInfoFromDomain(ctx context.Context, d *seed.Domain) (domainInfo, error) {
	objectMap := make(map[seed.CodeName]objectInfo, len(d.Objects))
	var err error
	for _, ob := range d.Objects {
		objectMap[ob.Name], err = db.objectInfoFromObject(ctx, d, ob)
		if err != nil {
			return domainInfo{}, err
		}
	}
	return domainInfo{
		Thing:     d.Thing.DeepCopy(),
		objectMap: objectMap,
	}, nil
}

type objectInfo struct {
	seed.Thing
	fields       *orderedmap.OrderedMap[seed.CodeName, fieldInfo]
	mainTable    Table
	helperTables map[string]Table
}

func (db *DB) objectInfoFromObject(ctx context.Context, d *seed.Domain, ob seed.Object) (objectInfo, error) {
	tableName, err := db.generateTableName(ctx, d, &ob)
	if err != nil {
		return objectInfo{}, err
	}
	fields := orderedmap.New[seed.CodeName, fieldInfo]()
	table := InitTable(tableName)
	table.Option = db.option.TableOption
	helpers := make(map[string]Table)
	for i := range ob.Fields {
		f := &ob.Fields[i]
		info, err := db.generateFieldInfo(f)
		if err != nil {
			return objectInfo{}, err
		}
		_, present := fields.Set(f.Name, *info)
		if present {
			return objectInfo{}, seederrors.NewSystemError(`field with name="%s" inserted again, this should never happen`, f.Name)
		}
		for _, col := range info.cols {
			_, present = table.Columns.Set(col.Name, col)
			if present {
				return objectInfo{}, seederrors.NewSystemError(`column with name="%s" inserted again, this should never happen`, col.Name)
			}
		}
		table.Constraint.Checks = append(table.Constraint.Checks, info.checks...)
		for _, helper := range info.tables {
			_, present = helpers[helper.Name]
			if present {
				return objectInfo{}, seederrors.NewSystemError(`table with name="%s" inserted again, this should never happen`, helper.Name)
			}
			helpers[helper.Name] = helper
		}
	}
	return objectInfo{
		Thing:        ob.Thing.DeepCopy(),
		fields:       fields,
		mainTable:    table,
		helperTables: helpers,
	}, nil
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
