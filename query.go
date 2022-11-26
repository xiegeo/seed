package seed

type ObjectQuery struct {
	ObjetName CodeName
	Fields    NameTree
	Condition Condition
	Count     bool
	Order     []PartialOrder
	Offset    int64
	Limit     int64 // future: use order and last row data for offset condition
}

type NameTree map[CodeName]NameTree

type PartialOrder struct {
	FieldPath []CodeName
	Inverse   bool
}
