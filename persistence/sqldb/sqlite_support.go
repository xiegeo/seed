package sqldb

import (
	"math"
	"math/big"

	"github.com/xiegeo/seed"
)

const (
	sqliteTableDefinition       = "STRICT"
	sqliteHelperTableDefinition = "STRICT, WITHOUT ROWID"
)

func Sqlite(op *DBOption) error {
	op.ColumnFeatures = SqliteColumnFeatures()
	op.TableOption = sqliteTableDefinition
	return nil
}

func SqliteColumnFeatures() ColumnFeatures {
	maxBlobSize := int64(1_000_000_000) // default setting for SQLITE_MAX_LENGTH
	maxBlobSize = maxBlobSize / 2       // add a safety buffer
	var cs ColumnFeatures
	cs.Append("TEXT", false, &seed.Field{
		FieldType: seed.String,
		FieldTypeSetting: seed.StringSetting{
			MaxCodePoints: maxBlobSize / 4, // a code point is at most 4 bytes
		},
	})
	cs.Append("BLOB", false, &seed.Field{
		FieldType: seed.Binary,
		FieldTypeSetting: seed.BinarySetting{
			MaxBytes: maxBlobSize,
		},
	})
	cs.Append("INTEGER", false, &seed.Field{
		FieldType: seed.Integer,
		FieldTypeSetting: seed.IntegerSetting{
			Min: big.NewInt(math.MinInt64),
			Max: big.NewInt(math.MaxInt64),
		},
	})
	cs.Append("REAL", false, &seed.Field{
		FieldType: seed.Real,
		FieldTypeSetting: seed.RealSetting{
			Standard: seed.Float64,
		},
	})
	return cs
}
