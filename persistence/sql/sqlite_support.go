package sql

import (
	"math"
	"math/big"

	"github.com/xiegeo/seed"
)

const (
	sqliteTableDefinition       = "STRICT"
	sqliteHelperTableDefinition = "STRICT, WITHOUT ROWID"
)

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

	/*
		switch f.FieldType {
		case seed.String:
			c.Type = "TEXT"
		case seed.Binary:
			c.Type = "BLOB"
		case seed.Boolean:
			c.Type = "INTEGER" // 0: false, 1: true
		case seed.TimeStamp:
			setting, err := seed.GetFieldTypeSetting[seed.TimeStampSetting](f)
			if err != nil {
				return err
			}
			if setting.WithTimeZone {
				return seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name, "WithTimeZone")
			}
			c.Type = "TEXT"
		case seed.Integer:
			c.Type = "INTEGER"
		case seed.Real:
			c.Type = "REAL" // float64
		default:
			return seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name)
		}
		return nil
	*/
}
