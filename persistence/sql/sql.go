package sql

import (
	"io"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/xiegeo/seed/seederrors"
)

type Table struct {
	Name       string
	Columns    *orderedmap.OrderedMap[string, Column]
	Constraint TableConstraint
	Option     TableOption
}

type CreateTable Table

func (t CreateTable) WriteTo(w io.Writer) (int64, error) {
	warpper := newWriteWarpper(w)
	t.writeTo(warpper)
	return warpper.n, warpper.err
}

func (t CreateTable) writeTo(w *writeWarpper) {
	if t.Columns.Len() == 0 {
		w.additionalError(seederrors.NewFieldsNotDefinedError(t.Name))
	}
	w.printf("CREATE TABLE %s (\n\t", t.Name)
	column := t.Columns.Oldest()
	column.Value.writeTo(w)
	for ; column != nil; column = column.Next() {
		w.printf(",\n\t")
		column.Value.writeTo(w)
	}
	t.Constraint.writeTo(w)
	w.printf("\n) %s;", t.Option)
}

type Column struct {
	Name       string
	Type       ColumnType
	TypeArg    []int64
	Constraint ColumnConstraint
}

func (c Column) writeTo(w *writeWarpper) {
	// todo
}

type ColumnType string

type ColumnConstraint struct {
	NotNull bool
}

type TableConstraint struct {
	PrimaryKeys []string
	Uniques     [][]string
	Checks      []Expression
}

func (c TableConstraint) writeTo(w *writeWarpper) {
	// todo
}

type Expression struct {
	Type        ExpressionType
	A           string // a literal value, operator, or function name
	Expressions []Expression
}

type ExpressionType int8

const (
	ValueExpression ExpressionType = iota
	UnaryExpression
	BinaryExpression
)

type TableOption string
