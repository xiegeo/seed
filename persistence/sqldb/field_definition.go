package sqldb

import (
	"fmt"
	"math/big"
	"time"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

type fieldDefinition struct {
	cols   []Column
	checks []Expression
	tables []Table
}

func (d fieldDefinition) append(d2 fieldDefinition) fieldDefinition {
	d.cols = append(d.cols, d2.cols...)
	d.checks = append(d.checks, d2.checks...)
	d.tables = append(d.tables, d2.tables...)
	return d
}

// generateFieldDefinition support a seed defined field with 0 to many columns and 0 to many tables.
// of both columns and tables are none, an error must be returned.
func (db *DB) generateFieldDefinition(f *seed.Field) (fieldDefinition, error) {
	if f.IsI18n {
		return fieldDefinition{}, seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name, "IsI18n")
	}
	col, found := db.option.ColumnFeatures.Match(f)
	if found {
		return col.fieldDefinition(f)
	}
	switch setting := f.FieldTypeSetting.(type) {
	case seed.BooleanSetting:
		return db.generateFieldDefinition(boolAsIntegerField(f))
	case seed.TimeStampSetting:
		return db.timeStampFailback(f, setting)
	case nil:
		return fieldDefinition{}, fmt.Errorf("FieldTypeSetting not set")
	}
	return fieldDefinition{}, seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name)
}

// boolAsIntegerField implements boolean values using 0 for false, 1 for true.
func boolAsIntegerField(f *seed.Field) *seed.Field {
	out := *f
	out.FieldType = seed.Integer
	out.FieldTypeSetting = seed.IntegerSetting{
		Min: big.NewInt(0),
		Max: big.NewInt(1),
	}
	return &out
}

// timeZoneOffsetField describes a time zone offset field using integer seconds.
func timeZoneOffsetField(base *seed.Field) *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: base.Name + "_tz",
		},
		FieldType: seed.Integer,
		FieldTypeSetting: seed.IntegerSetting{
			Min: big.NewInt(int64(-24 * time.Hour / time.Second)),
			Max: big.NewInt(int64(24 * time.Hour / time.Second)),
		},
	}
}

var _failbackTimeStampCoverage = seed.TimeStampSetting{
	Min:   time.Time{},                                   // Min uses go zero time
	Max:   time.Date(10000, 1, 1, 0, 0, 0, -1, time.UTC), // 9999 time stamp
	Scale: time.Nanosecond,
}

func (db *DB) timeStampFailback(f *seed.Field, setting seed.TimeStampSetting) (fieldDefinition, error) {
	if !setting.WithTimeZoneOffset { // use strings if time stamp without time zone is not supported natively.
		if !_failbackTimeStampCoverage.Covers(setting) {
			return fieldDefinition{}, seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name, "FieldTypeSetting")
		}
		utcField := *f
		utcField.FieldType = seed.String
		utcField.FieldTypeSetting = seed.StringSetting{
			MinCodePoints: 1,                                           // the empty string is not allowed
			MaxCodePoints: int64(len("yyyy-mm-ddThh:mm:ss.123456789")), // enough for any value from year 1 to 9999 in nano seconds
		}
		return db.generateFieldDefinition(&utcField)
	}
	setting.WithTimeZoneOffset = false
	f2 := *f
	f2.FieldTypeSetting = setting
	utcTime, err := db.generateFieldDefinition(&f2) // don't call timeStampFailback directly because time stamp without time zone could be supported natively.
	if err != nil {
		return fieldDefinition{}, err
	}
	timeZoneOffset, err := db.generateFieldDefinition(timeZoneOffsetField(f))
	if err != nil {
		return fieldDefinition{}, err
	}
	return utcTime.append(timeZoneOffset), nil
}
