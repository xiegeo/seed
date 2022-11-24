package sqldb

import (
	"context"
	"reflect"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

// InsertObjects insert data keyed by object code name. If value is a slice, it's treated as a list of values.
func (db *DB) InsertObjects(ctx context.Context, v map[seed.CodeName]any) error {
	bt := newBatchTables(db.defaultDomain)
	for name, value := range v {
		err := ctx.Err()
		if err != nil {
			return err
		}
		err = bt.appendData(name, value)
		if err != nil {
			return err
		}
	}
	return db.doTransaction(ctx, func(txc txContext) error {
	})
}

type batchTables struct {
	domain domainInfo
	tables map[string]batchRows
}

type batchRows struct {
	columnIndexes map[string]int
	rows          [][]any // [rows][cols]value
}

func newBatchTables(domain domainInfo) *batchTables {
	return &batchTables{
		domain: domain,
		tables: make(map[string]batchRows),
	}
}

func (b *batchTables) getTableRows(obInfo *objectInfo) batchRows {
	rowValues := b.tables[obInfo.mainTable.Name]
	if len(rowValues.columnIndexes) == 0 {
		rowValues.columnIndexes = obInfo.mainTable.ColumnIndexes()
	}
	return rowValues
}

func (b *batchTables) appendData(objectName seed.CodeName, data any) error {
	obInfo, ok := b.domain.objectMap[objectName]
	if !ok {
		return seederrors.NewObjectNotFoundError(objectName)
	}
	if data == nil {
		return nil
	}
	return b.appendValue(&obInfo, reflect.ValueOf(data))
}

func (b *batchTables) appendValue(obInfo *objectInfo, dataValue reflect.Value) error {
	dataValue, isNil := getElem(dataValue)
	if isNil {
		return nil
	}
	if !dataValue.IsValid() {
		return seederrors.NewSystemError(`reflected value "%s" is not valid`, dataValue)
	}
	switch dataValue.Kind() {
	default:
		return seederrors.NewSystemError("Kind %s in data of type %s not handled", dataValue.Kind(), dataValue.Type())
	case reflect.Array, reflect.Slice:
		for i := 0; i < dataValue.Len(); i++ {
			err := b.appendValue(obInfo, dataValue.Index(i))
			if err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		mapAsInterface := dataValue.Interface()
		mapTyped, ok := mapAsInterface.(map[string]any)
		if ok {
			return b.appendMapValue(obInfo, mapTyped)
		}
		return seederrors.NewSystemError("map of type %s is not handled, only map[string]any{} is supported", dataValue.Type())
	}
}

func (b *batchTables) appendMapValue(obInfo *objectInfo, m map[string]any) error {
	table := b.getTableRows(obInfo)
	row := make([]any, 0, len(table.columnIndexes))
	for current := obInfo.fields.Oldest(); current != nil; current = current.Next() {
		fieldName := current.Key
		fieldValue := m[string(fieldName)]
		valueColumns := current.Value.cols
		if current.Value.Nullable && isNilPointer(fieldValue) {
			row = append(row, make([]any, len(valueColumns))...) // fill the column of this value with nils
			continue
		}
		values, err := current.Value.encoder(fieldValue)
		if err != nil {
			return err
		}
		row = append(row, values...)
	}
	if len(row) != len(table.columnIndexes) {
		return seederrors.NewSystemError("can not set %d values to %d columns", len(row), len(table.columnIndexes))
	}
	table.rows = append(table.rows, row)
	return nil
}

// getElem removes all interface and pointer wrappers. If value ends in nil pointer, true is returned
func getElem(v reflect.Value) (reflect.Value, bool) {
	for {
		switch v.Kind() {
		case reflect.Interface, reflect.Pointer:
			if v.IsNil() {
				return v, true
			}
			v = v.Elem()
		default:
			return v, false
		}
	}
}

func isNilPointer(a any) bool {
	if a == nil {
		return true
	}
	_, isNil := getElem(reflect.ValueOf(a))
	return isNil
}
