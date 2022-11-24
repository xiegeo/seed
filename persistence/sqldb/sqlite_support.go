package sqldb

import (
	"math"
	"math/big"

	"github.com/xiegeo/seed"
)

const (
	sqliteTableDefinition       = "STRICT"
	sqliteHelperTableDefinition = "STRICT, WITHOUT ROWID"

	default_SQLITE_MAX_LENGTH = 1_000_000_000                 // default setting for SQLITE_MAX_LENGTH
	sqliteMaxBlobSize         = default_SQLITE_MAX_LENGTH / 2 // add a safety buffer
	maxCodePointSize          = 4                             // a code point is at most 4 bytes
)

func Sqlite(op *DBOption) error {
	op.ColumnFeatures = SqliteColumnFeatures()
	op.TableOption = sqliteTableDefinition
	return nil
}

func SqliteColumnFeatures() (features ColumnFeatures) {
	features.MustAppend("TEXT", false, &seed.Field{
		FieldType: seed.String,
		FieldTypeSetting: seed.StringSetting{
			MaxCodePoints: sqliteMaxBlobSize / maxCodePointSize,
		},
	})
	features.MustAppend("BLOB", false, &seed.Field{
		FieldType: seed.Binary,
		FieldTypeSetting: seed.BinarySetting{
			MaxBytes: sqliteMaxBlobSize,
		},
	})
	features.MustAppend("INTEGER", false, &seed.Field{
		FieldType: seed.Integer,
		FieldTypeSetting: seed.IntegerSetting{
			Min: big.NewInt(math.MinInt64),
			Max: big.NewInt(math.MaxInt64),
		},
	})
	features.MustAppend("REAL", false, &seed.Field{
		FieldType: seed.Real,
		FieldTypeSetting: seed.RealSetting{
			Standard: seed.Float64,
		},
	})
	return features
}
