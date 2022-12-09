package testdomain

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"golang.org/x/text/language"

	"github.com/xiegeo/must"
	"github.com/xiegeo/seed"
)

func TextLineField() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "text_10",
			Label: seed.I18n[string]{
				language.English: "Single Line Text",
			},
		},
		FieldType: seed.String,
		FieldTypeSetting: seed.StringSetting{
			MinCodePoints: 0,
			MaxCodePoints: 10,
			IsSingleLine:  true,
		},
	}
}

func TextAreaField() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "text_area_3000",
			Label: seed.I18n[string]{
				language.English: "Multiline Text",
			},
		},
		FieldType: seed.String,
		FieldTypeSetting: seed.StringSetting{
			MinCodePoints: 0,
			MaxCodePoints: 3000,
			IsSingleLine:  false,
		},
	}
}

func Bytes() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "bytes_10",
			Label: seed.I18n[string]{
				language.English: "Binary Data",
			},
		},
		FieldType: seed.Binary,
		FieldTypeSetting: seed.BinarySetting{
			MinBytes: 0,
			MaxBytes: 10,
		},
	}
}

func Bool() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "bool",
			Label: seed.I18n[string]{
				language.English: "True or False",
			},
		},
		FieldType:        seed.Boolean,
		FieldTypeSetting: seed.BooleanSetting{},
	}
}

const dayDuration = 24 * time.Hour

func Date() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "date_9999",
			Label: seed.I18n[string]{
				language.English: "Date",
			},
		},
		FieldType: seed.TimeStamp,
		FieldTypeSetting: seed.TimeStampSetting{
			Min:                time.Time{},                                   // Min uses go zero time
			Max:                time.Date(10000, 1, 1, 0, 0, 0, -1, time.UTC), // 9999 time stamp
			Scale:              dayDuration,
			WithTimeZoneOffset: false,
		},
	}
}

func DateTimeSec() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "datetime_sec_9999",
			Label: seed.I18n[string]{
				language.English: "Date and Time",
			},
		},
		FieldType: seed.TimeStamp,
		FieldTypeSetting: seed.TimeStampSetting{
			Min:                time.Time{},                                   // Min uses go zero time
			Max:                time.Date(10000, 1, 1, 0, 0, 0, -1, time.UTC), // 9999 time stamp
			Scale:              time.Second,
			WithTimeZoneOffset: false,
		},
	}
}

func DateTimeMill() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "datetime_mill_9999",
			Label: seed.I18n[string]{
				language.English: "3 Decimal Seconds",
			},
		},
		FieldType: seed.TimeStamp,
		FieldTypeSetting: seed.TimeStampSetting{
			Min:                time.Time{},                                   // Min uses go zero time
			Max:                time.Date(10000, 1, 1, 0, 0, 0, -1, time.UTC), // 9999 time stamp
			Scale:              time.Millisecond,
			WithTimeZoneOffset: false,
		},
	}
}

// DateTimeMicro is what most SQL implantation support natively
func DateTimeMicro() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "datetime_micro_9999",
			Label: seed.I18n[string]{
				language.English: "6 Decimal Seconds",
			},
		},
		FieldType: seed.TimeStamp,
		FieldTypeSetting: seed.TimeStampSetting{
			Min:                time.Time{},                                   // Min uses go zero time
			Max:                time.Date(10000, 1, 1, 0, 0, 0, -1, time.UTC), // 9999 time stamp
			Scale:              time.Microsecond,
			WithTimeZoneOffset: false,
		},
	}
}

func DateTimeNano() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "datetime_nano_9999",
			Label: seed.I18n[string]{
				language.English: "9 Decimal Seconds",
			},
		},
		FieldType: seed.TimeStamp,
		FieldTypeSetting: seed.TimeStampSetting{
			Min:                time.Time{},                                   // Min uses go zero time
			Max:                time.Date(10000, 1, 1, 0, 0, 0, -1, time.UTC), // 9999 time stamp
			Scale:              time.Nanosecond,
			WithTimeZoneOffset: false,
		},
	}
}

func WithTimeZone(f *seed.Field) *seed.Field {
	setting := must.V(seed.GetFieldTypeSetting[seed.TimeStampSetting](f))
	setting.WithTimeZoneOffset = true
	return &seed.Field{
		Thing: seed.Thing{
			Name: "zoned_" + f.Thing.Name,
			Label: seed.I18n[string]{
				language.English: f.Thing.Label[language.English] + " With Time Zone",
			},
		},
		FieldType:        seed.TimeStamp,
		FieldTypeSetting: setting,
	}
}

func PositiveInteger64() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "positive_integer_64",
			Label: seed.I18n[string]{
				language.English: "Number 1 to Max i64",
			},
		},
		FieldType: seed.Integer,
		FieldTypeSetting: seed.IntegerSetting{
			Min: big.NewInt(1),
			Max: big.NewInt(math.MaxInt64),
		},
	}
}

func Integer64() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "integer_64",
			Label: seed.I18n[string]{
				language.English: "Number (i64)",
			},
		},
		FieldType: seed.Integer,
		FieldTypeSetting: seed.IntegerSetting{
			Min: big.NewInt(math.MinInt64),
			Max: big.NewInt(math.MaxInt64),
		},
	}
}

const f64MaxSafeInterger = 9007199254740991 // the maximum safe integer in JavaScript (2^53 – 1), which uses f64 to represent all numbers.

func JSInteger() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "integer_js",
			Label: seed.I18n[string]{
				language.English: "JS safe integer",
			},
		},
		FieldType: seed.Integer,
		FieldTypeSetting: seed.IntegerSetting{
			Min: big.NewInt(-f64MaxSafeInterger),
			Max: big.NewInt(f64MaxSafeInterger),
		},
	}
}

func BigMax(f *seed.Field, exp int64) *seed.Field {
	setting := must.V(seed.GetFieldTypeSetting[seed.IntegerSetting](f))
	max := big.NewInt(0).Exp(big.NewInt(math.MaxInt64), big.NewInt(exp), nil)
	setting.Max = max
	return &seed.Field{
		Thing: seed.Thing{
			Name: seed.CodeName(fmt.Sprintf("max64e%d_%s", exp, f.Thing.Name)),
			Label: seed.I18n[string]{
				language.English: fmt.Sprintf("Max(%d) %s", exp, f.Thing.Label[language.English]),
			},
		},
		FieldType:        seed.TimeStamp,
		FieldTypeSetting: setting,
	}
}

func Float64() *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: "float_64",
			Label: seed.I18n[string]{
				language.English: "Real (f64)",
			},
		},
		FieldType: seed.Real,
		FieldTypeSetting: seed.RealSetting{
			Standard: seed.Float64,
		},
	}
}

func ListOf(f *seed.Field, listSetting seed.ListSetting) *seed.Field {
	listSetting.ItemType = f.FieldType
	listSetting.ItemTypeSetting = f.FieldTypeSetting
	return &seed.Field{
		Thing: seed.Thing{
			Name: "list_" + f.Thing.Name,
			Label: seed.I18n[string]{
				language.English: f.Thing.Label[language.English] + " List",
			},
		},
		FieldType:        seed.List,
		FieldTypeSetting: listSetting,
	}
}
