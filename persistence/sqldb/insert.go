package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

// InsertObjects insert data keyed by object code name. If value is a slice, it's treated as a list of values.
// All inserts must be done or none at all.
func (db *DB) InsertObjects(ctx context.Context, v map[seed.CodeName]any) error {
	return db.InsertDomainObjects(ctx, db.defaultDomain, v)
}

func (db *DB) InsertDomainObjects(ctx context.Context, domain *domainInfo, v map[seed.CodeName]any) error {
	batch := newBatchTables(domain)
	for name, value := range v {
		err := ctx.Err()
		if err != nil {
			return err
		}
		err = batch.appendData(name, value)
		if err != nil {
			return err
		}
	}
	return db.doTransaction(ctx, func(txc txContext) error {
		for tableName, tableContent := range batch.tables {
			if len(tableContent.rows) == 0 {
				continue
			}
			stmt, err := tableContent.insertRowStmt(txc, tableName, db.option.TranslateStatement)
			if err != nil {
				return err
			}
			for _, row := range tableContent.rows {
				_, err = stmt.Exec(row...)
				if err != nil {
					return err
				}
			}
		}
		return nil
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

func (b *batchRows) insertRowStmt(txc txContext, tableName string, q func(string) string) (*sql.Stmt, error) {
	colNames := make([]string, len(b.columnIndexes))
	for name, i := range b.columnIndexes {
		if colNames[i] != "" {
			return nil, seederrors.NewSystemError("table %s columnIndexes %v must be a 1 to 1 map, got repeated %d", tableName, b.columnIndexes, i)
		}
		colNames[i] = name
	}
	return txc.Prepare(q(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName, strings.Join(colNames, ", "), joinNth("?", ",", len(colNames)))))
}

func joinNth(s, sep string, n int) string {
	slice := make([]string, n)
	for i := range slice {
		slice[i] = s
	}
	return strings.Join(slice, sep)
}

func newBatchTables(domain *domainInfo) *batchTables {
	return &batchTables{
		domain: *domain,
		tables: make(map[string]batchRows),
	}
}

func (b *batchTables) getTableRows(obInfo *objectInfo) batchRows {
	rowValues := b.tables[obInfo.mainTable.TableName()]
	if len(rowValues.columnIndexes) == 0 {
		rowValues.columnIndexes = obInfo.mainTable.ColumnIndexes()
	}
	return rowValues
}

func (b *batchTables) appendData(objectName seed.CodeName, data any) error {
	obInfo, ok := b.domain.objectMap.Get(objectName)
	if !ok {
		return seederrors.NewObjectNotFoundError(objectName)
	}
	if data == nil {
		return nil
	}
	return b.appendValue(obInfo, reflect.ValueOf(data))
}

func (b *batchTables) appendValue(obInfo *objectInfo, dataValue reflect.Value) error {
	dataValue, isNil := getElem(dataValue)
	if isNil {
		return nil
	}
	if !dataValue.IsValid() {
		return seederrors.NewSystemError(`reflected value (%s) is not valid`, dataValue)
	}
	switch dataValue.Kind() {
	default:
		return seederrors.NewSystemError("Kind %s in input of type %s not handled, use map for single data and slice for batch, structs will be supported in the future", dataValue.Kind(), dataValue.Type())
	case reflect.Array, reflect.Slice:
		for i := 0; i < dataValue.Len(); i++ {
			err := b.appendValue(obInfo, dataValue.Index(i))
			if err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		switch mapTyped := dataValue.Interface().(type) {
		case map[string]any:
			return appendMapValue(b, obInfo, mapTyped)
		case map[seed.CodeName]any:
			return appendMapValue(b, obInfo, mapTyped)
		}
		return seederrors.NewSystemError("map of type %s is not handled, only map[string|seed.CodeName]any{} is supported", dataValue.Type())
	}
}

func appendMapValue[K ~string](b *batchTables, obInfo *objectInfo, m map[K]any) error {
	table := b.getTableRows(obInfo)
	row := make([]any, 0, len(table.columnIndexes))
	err := obInfo.fields.RangeLogical(func(fieldName seed.CodeName, fi *fieldInfo) error {
		fieldValue := m[K(fieldName)]
		valueColumns := fi.cols
		if isNilPointer(fieldValue) {
			if fi.Nullable {
				row = append(row, make([]any, len(valueColumns))...) // fill the columns of this value with nils
				return nil
			}
			return seederrors.NewValueRequiredError(fieldName)
		}
		values, err := fi.Encoder()(fieldValue)
		if err != nil {
			return err
		}
		row = append(row, values...)
		return nil
	})
	if err != nil {
		return err
	}
	if len(row) != len(table.columnIndexes) {
		return seederrors.NewSystemError("can not set %d values to %d columns", len(row), len(table.columnIndexes))
	}
	table.rows = append(table.rows, row)
	b.tables[obInfo.mainTable.TableName()] = table
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
