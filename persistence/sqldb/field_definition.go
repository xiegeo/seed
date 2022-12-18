package sqldb

import (
	"math/big"
	"time"

	"github.com/xiegeo/must"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

// fieldDefinition support a seed defined field with 0 to many columns and 0 to many tables.
type fieldDefinition struct {
	// sql definitions
	cols   []Column
	checks []Expression
	tables []*Table

	// for operations
	eqColumns         []string // Columns used for field equality check. If empty, defaults to all.
	sortColumns       []string // Columns and order used for comparisons. If empty, defaults to Column natural order.
	invertSortColumns []string // Columns that have invers ordering, such as time zone offset.
}

// appendNoEq append all definitions except eqColumns
func (d fieldDefinition) appendNoEq(d2 fieldDefinition) fieldDefinition {
	d.cols = append(d.cols, d2.cols...)
	d.checks = append(d.checks, d2.checks...)
	d.tables = append(d.tables, d2.tables...)
	d.sortColumns = append(d.sortColumns, d2.sortColumns...)
	d.invertSortColumns = append(d.invertSortColumns, d2.invertSortColumns...)

	d.eqColumns = d.getEqColumns()
	return d
}

func GetColumnNames(cs []Column) []string {
	names := make([]string, len(cs))
	for i, c := range cs {
		names[i] = c.Name
	}
	return names
}

func (d fieldDefinition) getEqColumns() []string {
	if len(d.eqColumns) > 0 {
		return d.eqColumns
	}
	return GetColumnNames(d.cols)
}

func (d fieldDefinition) getSortColumns() []string {
	if len(d.sortColumns) > 0 {
		return d.sortColumns
	}
	return GetColumnNames(d.cols)
}

// generateFieldInfo need table to have Name and Constraint.PrimaryKeys set, if any.
// So that the table can be referenced
func (builder *fieldInfoBuilder) generateFieldInfo(f *seed.Field) (*fieldInfo, error) {
	fi, err := builder.generateFieldInfoSub(f)
	if err != nil {
		return nil, err
	}
	fi.Field = *f
	return fi, nil
}

// generateFieldInfoSub is generateFieldInfo without setting the fieldInfo.Field to match input
func (builder *fieldInfoBuilder) generateFieldInfoSub(f *seed.Field) (*fieldInfo, error) {
	if f.IsI18n {
		return nil, seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name, "IsI18n")
	}
	col, found := builder.db.option.ColumnFeatures.Match(f)
	if found {
		fd, err := col.fieldDefinition(f)
		return &fieldInfo{
			Field:           *f,
			fieldDefinition: fd,
		}, err
	}
	switch setting := f.FieldTypeSetting.(type) {
	case seed.BooleanSetting:
		fd, err := builder.generateFieldInfoSub(boolAsIntegerField(f))
		fd.Field = *f // return the original boolean field, sql supports casting bool to 0 and 1
		return fd, err
	case seed.TimeStampSetting:
		return builder.timeStampFailback(f, setting)
	case seed.ListSetting:
		return builder.listFieldInfo(f, setting)
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

const day = 24 * time.Hour

func utcTimeLayoutForScale(scale time.Duration) string {
	if scale%day == 0 { //nolint:gocritic // if else represents condition ordering much more strongly than switch case
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

func (builder *fieldInfoBuilder) utcTimeStampFailback(f *seed.Field, setting seed.TimeStampSetting) (*fieldInfo, error) {
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
	fi, err := builder.generateFieldInfoSub(&utcField)
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
	return fi, nil
}

func (builder *fieldInfoBuilder) timeStampFailback(f *seed.Field, setting seed.TimeStampSetting) (*fieldInfo, error) {
	if !setting.WithTimeZoneOffset {
		// use strings if time stamp without time zone is not supported natively.
		return builder.utcTimeStampFailback(f, setting)
	}
	setting.WithTimeZoneOffset = false
	f2 := *f
	f2.FieldTypeSetting = setting
	utcTime, err := builder.generateFieldInfoSub(&f2) // don't call utcTimeStampFailback directly because time stamp without time zone could be supported natively.
	if err != nil {
		return nil, err
	}
	timeZoneOffset, err := builder.generateFieldInfoSub(timeZoneOffsetField(f))
	if err != nil {
		return nil, err
	}
	timeZoneOffset.invertSortColumns = GetColumnNames(timeZoneOffset.cols)
	return timeStampFailback(utcTime, timeZoneOffset), nil
}

func timeStampFailback(utcTime, timeZoneOffset *fieldInfo) *fieldInfo {
	encodeUTC := utcTime.Encoder()
	encodeOffset := timeZoneOffset.Encoder()
	decodeUTC := utcTime.Decoder()
	decodeOffset := timeZoneOffset.Decoder()
	fd := utcTime.fieldDefinition.appendNoEq(timeZoneOffset.fieldDefinition) // equal for timestamp ignores time zone
	return &fieldInfo{
		fieldDefinition: fd,
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
			offsetTyped, ok := offsetInter.(int64)
			if !ok {
				return nil, seederrors.NewSystemError("decoder expected int64 for time zone offset, but got %T", offsetInter)
			}
			return utcTyped.In(time.FixedZone("", int(offsetTyped))), nil
		},
	}
}

func (builder *fieldInfoBuilder) listFieldInfo(f *seed.Field, setting seed.ListSetting) (*fieldInfo, error) {
	table := builder.initHelperTable(f)
	parentPK := builder.parent.PrimaryKeys()
	if len(parentPK) == 0 {
		return nil, seederrors.NewSystemError("parent table must have primary key defined")
	}
	forgeignKey := ForeignKey[*TableName]{
		Keys:       withPrefix(systemColumnPrefix, parentPK...),
		TableName:  builder.parent.Name,
		References: parentPK,
		OnDelete:   OnActionCascade,
		OnUpdate:   OnActionCascade,
	}
	table.Constraint.ForeignKeys = append(table.Constraint.ForeignKeys, forgeignKey)
	rev := ExternalColumnName("").Revert
	for _, col := range forgeignKey.Keys {
		_, _ = must.B2(table.Columns.Set(rev(col), Column{
			Name:       col,
			Constraint: ColumnConstraint{NotNull: true},
		}))(must.Any, false, "must not over write key")
	}
	elem, err := builder.generateFieldInfoSub(setting.ItemField(f))
	if err != nil {
		return nil, seederrors.WithMessagef(err, "listFieldInfo generate item field for %s", f.Name)
	}
	order, err := builder.generateFieldInfoSub(listOrderField(f, setting))
	if err != nil {
		return nil, seederrors.WithMessagef(err, "listFieldInfo generate list order for %s", f.Name)
	}
	localKeys := order.getEqColumns()
	table.Constraint.PrimaryKeys = append(forgeignKey.Keys, localKeys...) //nolint:gocritic // append result not assigned to the same slice
	for _, col := range append(elem.cols, order.cols...) {
		_, _ = must.B2(table.Columns.Set(rev(col.Name), col))(must.Any, false, "must not over write key")
	}
	return &fieldInfo{
		Field: *f,
		fieldDefinition: fieldDefinition{
			tables: []*Table{table},
		},
	}, nil
}

func withPrefix[S ~string](prefix S, ss ...S) []S {
	for i, s := range ss {
		ss[i] = prefix + s
	}
	return ss
}

// listOrderField describes a _order field using integer.
func listOrderField(base *seed.Field, setting seed.ListSetting) *seed.Field {
	return &seed.Field{
		Thing: seed.Thing{
			Name: base.Name + systemColumnOrder,
		},
		FieldType: seed.Integer,
		FieldTypeSetting: seed.IntegerSetting{
			Min: big.NewInt(1),
			Max: big.NewInt(setting.MaxLength),
		},
	}
}
