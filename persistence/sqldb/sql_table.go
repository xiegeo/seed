package sqldb

import (
	"fmt"
	"io"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/xiegeo/must"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

const (
	systemColumnPrefix   = "_"                          // this prefix is not allowed by seed naming rules, so it's safe to use by internal columns
	systemColumnID       = systemColumnPrefix + "id"    // for auto increment key
	systemColumnTimeZone = systemColumnPrefix + "tz"    // for time zone offset
	systemColumnOrder    = systemColumnPrefix + "order" // for ordered list
	systemColumnCount    = systemColumnPrefix + "count" // for counted set
)

type TableName struct {
	Object    seed.ObjectNamePath
	FieldPath []seed.CodeName
	string    // caches String()
}

func (tn TableName) NewWithField(cn seed.CodeName) *TableName {
	tn.FieldPath = append(tn.FieldPath, cn)
	tn.string = ""
	return &tn
}

func (tn *TableName) String() string {
	if tn.string == "" {
		tn.string = fmt.Sprintf("%s_%s", tn.Object.Domain, tn.Object.Object)
		if len(tn.FieldPath) > 0 {
			tn.string += fmt.Sprintf("__%s", joinS(tn.FieldPath, "_"))
		}
	} else {
		if pathSize := len(tn.FieldPath); pathSize > 0 {
			must.True(strings.HasSuffix(tn.string, string(tn.FieldPath[pathSize-1])), "assert constant TableName")
		} else {
			must.True(strings.HasSuffix(tn.string, string(tn.Object.Object)), "assert constant TableName")
		}
	}
	return tn.string
}

func joinS[S ~string](elems []S, sep string) string {
	ss := make([]string, len(elems))
	for i, e := range elems {
		ss[i] = string(e)
	}
	return strings.Join(ss, sep)
}

// ATable describes an AST for a table, with conveniences to inspect certain table properties.
//
// ATable is designed so it could be extracted to a separate library unrelated to seed.
//
// Generic parameter T gives the table name.
//
// Generic parameter C gives the external column name. This allows users to use reserved
// SQL keywords as external column names that maps to a legal real column names. Any method
// which returns a C returns the external column name, and any method which returns string
// returns the raw expression.
type ATable[T fmt.Stringer, C ~string] struct {
	Name       T
	Columns    *orderedmap.OrderedMap[C, Column]
	Constraint TableConstraint[T]
	Option     TableOption
}

type Table = ATable[*TableName, ExternalColumnName]

// ExternalColumnName is either a seed.NameCode or system column.
// ExternalColumnName should never end in "_".
//
// Whether a word is reserved is database dependent, so that check can not be defined here.
type ExternalColumnName string

// Safe returns a string that is guaranteed not to be a reserved word. If the receiver is
// already not a reserved word, a direct cast is preferred to calling Safe.
func (n ExternalColumnName) Safe() string {
	return string(n) + "_"
}

// Revert reverts any raw column name back. It is tied to receiver so it could be part of
// an interface.
func (ExternalColumnName) Revert(s string) ExternalColumnName {
	if strings.HasSuffix(s, "_") {
		return ExternalColumnName(s[:len(s)-1])
	}
	return ExternalColumnName(s)
}

func (builder *objectInfoBuilder) initTable() *Table {
	return &Table{
		Name:    &TableName{Object: seed.ObjectNamePath{Domain: builder.domain.GetName(), Object: builder.object.GetName()}},
		Columns: orderedmap.New[ExternalColumnName, Column](),
	}
}

func (builder *fieldInfoBuilder) initHelperTable(f seed.ThingGetter) *Table {
	return &Table{
		Name:    builder.parent.Name.NewWithField(f.GetName()),
		Columns: orderedmap.New[ExternalColumnName, Column](),
	}
}

func (t *ATable[T, C]) TableName() string {
	return t.Name.String()
}

func (t *ATable[T, C]) ColumnIndexes() map[string]int {
	out := make(map[string]int, t.Columns.Len())
	var i int
	for column := t.Columns.Oldest(); column != nil; column = column.Next() {
		if strings.HasPrefix(string(column.Key), systemColumnPrefix) {
			continue // skip system prefix
		}
		out[column.Value.Name] = i
		i++
	}
	return out
}

func (t *ATable[T, C]) PrimaryKeys() []string {
	if len(t.Constraint.PrimaryKeys) > 0 {
		return t.Constraint.PrimaryKeys
	}
	return []string{systemColumnID} // this is a short-cut for our use case, should be collect all primary keys from column definitions.
}

type CreateTable struct {
	*Table
}

func MakeCreateTable(t *Table) CreateTable {
	return CreateTable{Table: t}
}

func (t CreateTable) WriteTo(w io.Writer) (int64, error) {
	warpper := newWriteWarpper(w)
	t.writeTo(warpper)
	return warpper.n, warpper.err
}

func (t CreateTable) writeTo(w *writeWarpper) {
	if t.Columns.Len() == 0 {
		w.additionalError(seederrors.NewFieldsNotDefinedError(t.TableName()))
	}
	w.printf("CREATE TABLE %s (\n\t", t.Name)
	column := t.Columns.Oldest()
	column.Value.writeTo(w)
	column = column.Next()
	for ; column != nil; column = column.Next() {
		w.printf(",\n\t")
		column.Value.writeTo(w)
	}
	t.Constraint.writeTo(w)
	w.printf("\n) %s;", strings.Join(t.Option, ", "))
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
	PrimaryKey bool
	NotNull    bool
}

func (c ColumnConstraint) writeTo(w *writeWarpper) {
	if c.PrimaryKey {
		w.write([]byte("PRIMARY KEY "))
	}
	if c.NotNull {
		w.write([]byte("NOT NULL"))
	}
}

type TableConstraint[T fmt.Stringer] struct {
	PrimaryKeys []string
	Uniques     [][]string
	ForeignKeys []ForeignKey[T]
	Checks      []Expression
}

func (c TableConstraint[T]) writeTo(w *writeWarpper) {
	if len(c.PrimaryKeys) > 0 {
		w.printf(",\n\tPRIMARY KEY (%s)", strings.Join(c.PrimaryKeys, ","))
	}
	for _, unique := range c.Uniques {
		w.printf(",\n\t     UNIQUE (%s)", strings.Join(unique, ","))
	}
	for _, fk := range c.ForeignKeys {
		w.printf(",\n\t")
		fk.writeTo(w)
	}
	for _, expression := range c.Checks {
		w.printf(",\n\t      CHECK (")
		expression.writeTo(w)
		w.printf(")")
	}
}

type ForeignKey[T fmt.Stringer] struct {
	Keys       []string
	TableName  T
	References []string
	OnDelete   OnAction
	OnUpdate   OnAction
}

type OnAction string

const (
	OnActionNoAction   OnAction = "NO ACTION"
	OnActionRestrict   OnAction = "RESTRICT"
	OnActionSetNull    OnAction = "SET NULL"
	OnActionSetDefault OnAction = "SET DEFAULT"
	OnActionCascade    OnAction = "CASCADE"
)

func (fk ForeignKey[T]) writeTo(w *writeWarpper) {
	w.printf("FOREIGN KEY (")
	w.printf(strings.Join(fk.Keys, ", "))
	w.printf(") REFERENCES ")
	w.write([]byte(fk.TableName.String()))
	w.printf("(%s)", strings.Join(fk.References, ", "))
	fk.OnDelete.writeToIfSet(w, "DELETE")
	fk.OnUpdate.writeToIfSet(w, "UPDATE")
}

func (a OnAction) writeToIfSet(w *writeWarpper, actionType string) {
	if a == "" {
		return
	}
	w.printf(" ON %s %s", actionType, a)
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

func ValueLiteral(v string) Expression {
	return Expression{
		Type: ValueExpression,
		A:    v,
	}
}

func (e Expression) writeTo(w *writeWarpper) {
	switch e.Type {
	default:
		w.additionalError(seederrors.NewSystemError("ExpressionType %d not handled", e.Type))
	case ValueExpression:
		w.write([]byte(e.A))
	case UnaryExpression:
		w.printf("%s ", e.A)
		if len(e.Expressions) != 1 {
			w.additionalError(seederrors.NewSystemError("UnaryExpression go %d Expressions", len(e.Expressions)))
		}
		e.Expressions[0].writeTo(w)
	case BinaryExpression:
		w.writeJoin([]byte(" "+e.A+" "), warpSlice(e.Expressions...)...)
	case ListExpression:
		w.writeArg(warpSlice(e.Expressions...)...)
	}
}

type TableOption []string

func (o TableOption) Add(opt ...string) TableOption {
	return append(o, opt...)
}
