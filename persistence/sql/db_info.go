package sql

import (
	"context"

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
	tableName string
	fieldList []seed.CodeName
	fieldMap  map[seed.CodeName]fieldInfo
}

func (db *DB) objectInfoFromObject(ctx context.Context, d *seed.Domain, ob seed.Object) (objectInfo, error) {
	tableName, err := db.generateTableName(ctx, d, &ob)
	if err != nil {
		return objectInfo{}, err
	}
	fieldList := make([]seed.CodeName, len(ob.Fields))
	fieldMap := make(map[seed.CodeName]fieldInfo, len(ob.Fields))
	for i, f := range ob.Fields {
		fieldList[i] = f.Name
		fieldMap[f.Name] = fieldInfo{
			Field: f.DeepCopy(),
		}
	}
	return objectInfo{
		Thing:     ob.Thing.DeepCopy(),
		tableName: tableName,
		fieldList: fieldList,
		fieldMap:  fieldMap,
	}, nil
}

type fieldInfo struct {
	seed.Field
	// colDef      []ColumnDefinition
	// helperTable TableDefinition
}
