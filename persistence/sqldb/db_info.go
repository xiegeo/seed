package sqldb

import (
	"context"
	"fmt"

	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/xiegeo/seed"
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
	for _, f := range ob.Fields {
		info := fieldInfo{
			Field: f,
		}
		info.fieldDefinition, err = db.generateFieldDefinition(&f)
		if err != nil {
			return objectInfo{}, err
		}
		_, present := fields.Set(f.Name, info)
		if present {
			return objectInfo{}, fmt.Errorf(`field with name="%s" inserted again, this should never happen`, f.Name)
		}
		for _, col := range info.cols {
			_, present = table.Columns.Set(col.Name, col)
			if present {
				return objectInfo{}, fmt.Errorf(`column with name="%s" inserted again, this should never happen`, col.Name)
			}
		}
		for _, check := range info.checks {
			table.Constraint.Checks = append(table.Constraint.Checks, check)
		}
		for _, helper := range info.tables {
			_, present = helpers[helper.Name]
			if present {
				return objectInfo{}, fmt.Errorf(`table with name="%s" inserted again, this should never happen`, helper.Name)
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
}
