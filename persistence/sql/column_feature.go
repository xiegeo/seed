package sql

type ColumnFeature struct {
	TypeName      ColumnType
	ArgRanges     []IntRange
	WriteNameOnly bool // if true, type args are not accepted by the database on column create (sqlite strict mode)
}

type IntRange struct {
	Min int64
	Max int64
}
