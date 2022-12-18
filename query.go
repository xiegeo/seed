package seed

type Query struct {
	ObjectName ObjectNamePath
	Fields     NameTree
	Condition  Condition
	Count      bool
	Order      []PartialOrder
	Offset     int64
	Limit      int64 // future: use order and last row data for offset condition
}

type ObjectNamePath struct {
	Domain CodeName
	Object CodeName
}

type NameTree map[CodeName]NameTree

type PartialOrder struct {
	FieldPath []CodeName
	Inverse   bool
}
