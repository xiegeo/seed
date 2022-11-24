package sqldb

import (
	"reflect"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

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
	dataValue = getElem(dataValue)
	if dataValue.IsNil() {
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
		if dataValue.Type().ConvertibleTo(reflect.TypeOf(map[string]any{})) {
			return b.appendMapValue(obInfo, dataValue.Interface().(map[string]any))
		}
		return seederrors.NewSystemError("map of type %s is not handled, only map[string]any{} is supported", dataValue.Type())
	}
}

func (b *batchTables) appendMapValue(obInfo *objectInfo, m map[string]any) error {
}

// getElem removes all interface and pointer wrappers
func getElem(v reflect.Value) reflect.Value {
	for {
		switch v.Kind() {
		case reflect.Interface, reflect.Pointer:
			if v.IsNil() {
				return v
			}
			v = v.Elem()
		default:
			return v
		}
	}
}
