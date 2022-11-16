package sql

import (
	"fmt"
	"io"
	"strings"

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
	TypeArg    []string
	Constraint ColumnConstraint
}

func (c Column) writeTo(w *writeWarpper) {
	w.printf("%s %s", c.Name, c.Type)
	if len(c.TypeArg) == 0 {
		w.printf(" ")
	} else {
		w.printf("(%s) ", strings.Join(c.TypeArg, ","))
	}
	c.Constraint.writeTo(w)
}

type ColumnType string

type ColumnConstraint struct {
	NotNull bool
}

func (c ColumnConstraint) writeTo(w *writeWarpper) {
	if c.NotNull {
		w.printf("NOT NULL")
	}
}

type TableConstraint struct {
	PrimaryKeys []string
	Uniques     [][]string
	Checks      []Expression
}

func (c TableConstraint) writeTo(w *writeWarpper) {
	if len(c.PrimaryKeys) > 0 {
		w.printf(",\n\tPRIMARY KEY (%s)", strings.Join(c.PrimaryKeys, ","))
	}
	for _, unique := range c.Uniques {
		w.printf(",\n\t     UNIQUE (%s)", strings.Join(unique, ","))
	}
	for _, expression := range c.Checks {
		w.printf(",\n\t      CHECK (")
		expression.writeTo(w)
		w.printf(")")
	}
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
	ListExpression
)

func (e Expression) writeTo(w *writeWarpper) {
	switch e.Type {
	default:
		w.additionalError(fmt.Errorf("ExpressionType %d not handled", e.Type))
	case ValueExpression:
		w.write([]byte(e.A))
	case UnaryExpression:
		w.printf("%s ", e.A)
		e.Expressions[0].writeTo(w)
	case BinaryExpression:
		w.writeJoin([]byte(" "+e.A+" "), warpSlice(e.Expressions...)...)
	case ListExpression:
		w.writeArg(warpSlice(e.Expressions...)...)
	}
}

type TableOption string
