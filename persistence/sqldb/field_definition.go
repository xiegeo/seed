package sqldb

import (
	"math/big"
	"time"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

// fieldDefinition support a seed defined field with 0 to many columns and 0 to many tables.
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

func (db *DB) generateFieldInfo(f *seed.Field) (*fieldInfo, error) {
	fi, err := db.generateFieldInfoSub(f)
	if err != nil {
		return nil, err
	}
	fi.Field = *f
	return fi, nil
}

// generateFieldInfoSub is generateFieldInfo without setting the fieldInfo.Field to match input
func (db *DB) generateFieldInfoSub(f *seed.Field) (*fieldInfo, error) {
	if f.IsI18n {
		return nil, seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name, "IsI18n")
	}
	col, found := db.option.ColumnFeatures.Match(f)
	if found {
		fd, err := col.fieldDefinition(f)
		return &fieldInfo{
			Field:           *f,
			fieldDefinition: fd,
		}, err
	}
	switch setting := f.FieldTypeSetting.(type) {
	case seed.BooleanSetting:
		fd, err := db.generateFieldInfoSub(boolAsIntegerField(f))
		fd.Field = *f // return the original boolean field, sql supports casting bool to 0 and 1
		return fd, err
	case seed.TimeStampSetting:
		return db.timeStampFailback(f, setting)
	case nil:
		return nil, seederrors.NewSystemError("FieldTypeSetting not set in field %s", f.Name)
	}
	return nil, seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name)
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

func utcTimeLayoutForScale(scale time.Duration) string {
	day := 24 * time.Hour
	if scale%day == 0 {
		return "2006-01-02"
	} else if scale%time.Second == 0 {
		return "2006-01-02T15:04:05"
	} else if scale%time.Millisecond == 0 {
		return "2006-01-02T15:04:05.000"
	} else if scale&time.Microsecond == 0 {
		return "2006-01-02T15:04:05.000000"
	}
	return "2006-01-02T15:04:05.000000000"
}

func utcTimeStringFuncForLayout(layout string) func(any) (any, error) {
	return func(v any) (any, error) {
		vt, ok := v.(time.Time)
		if !ok {
			return nil, seederrors.NewSystemError("encoder expect time.Time but got %T", v)
		}
		_, offset := vt.Zone()
		if offset != 0 {
			return nil, seederrors.NewSystemError("encoder expect time in UTC but got %s", vt.Location())
		}
		return vt.Format(layout), nil
	}
}

func (db *DB) timeStampFailback(f *seed.Field, setting seed.TimeStampSetting) (*fieldInfo, error) {
	if !setting.WithTimeZoneOffset { // use strings if time stamp without time zone is not supported natively.
		if !_failbackTimeStampCoverage.Covers(setting) {
			return nil, seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name, "FieldTypeSetting")
		}
		layout := utcTimeLayoutForScale(setting.Scale)
		codePoints := int64(len(layout))

		utcField := *f
		utcField.FieldType = seed.String
		utcField.FieldTypeSetting = seed.StringSetting{
			MinCodePoints: codePoints,
			MaxCodePoints: codePoints,
		}
		fi, err := db.generateFieldInfoSub(&utcField)
		if err != nil {
			return nil, err
		}
		fi.WarpEncoder(utcTimeStringFuncForLayout(layout))
		fi.WarpDecoder(func(a any) (any, error) {
			vt, ok := a.(string)
			if !ok {
				return nil, seederrors.NewSystemError("decoder expected string but got %T", a)
			}
			return time.Parse(layout, vt)
		})
	}
	setting.WithTimeZoneOffset = false
	f2 := *f
	f2.FieldTypeSetting = setting
	utcTime, err := db.generateFieldInfoSub(&f2) // don't call timeStampFailback directly because time stamp without time zone could be supported natively.
	if err != nil {
		return nil, err
	}
	timeZoneOffset, err := db.generateFieldInfoSub(timeZoneOffsetField(f))
	if err != nil {
		return nil, err
	}

	encodeUTC := utcTime.Encoder()
	encodeOffset := timeZoneOffset.Encoder()
	decodeUTC := utcTime.Decoder()
	decodeOffset := timeZoneOffset.Decoder()
	return &fieldInfo{
		fieldDefinition: utcTime.fieldDefinition.append(timeZoneOffset.fieldDefinition),
		encoder: func(v any) ([]any, error) {
			vt, ok := v.(time.Time)
			if !ok {
				return nil, seederrors.NewSystemError("encoder expect time.Time but got %T", v)
			}
			_, offset := vt.Zone()
			utcValue, err := encodeUTC(vt.UTC())
			if err != nil {
				return nil, err
			}
			offsetValue, err := encodeOffset(offset)
			if err != nil {
				return nil, err
			}
			return append(utcValue, offsetValue...), nil
		},
		decoder: func(a []any) (any, error) {
			utcValue := a[:len(utcTime.cols)]
			utcInter, err := decodeUTC(utcValue)
			if err != nil {
				return nil, err
			}
			utcTyped, ok := utcInter.(time.Time)
			if !ok {
				return nil, seederrors.NewSystemError("decoder expected time.Time but got %T", utcInter)
			}
			offsetValue := a[len(utcTime.cols):]
			offsetInter, err := decodeOffset(offsetValue)
			if err != nil {
				return nil, err
			}
		},
	}, nil
}
