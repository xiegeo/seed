package seedfake

import (
	"time"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

type ValueGen struct {
	NumberDistribution
	StringDistribution
	TimeDistribution
}

func NewValueGen(numbers NumberDistribution) *ValueGen {
	return &ValueGen{
		NumberDistribution: numbers,
		StringDistribution: NewRune(numbers, [2]rune{' ', '~'}),
		TimeDistribution:   NewTime(numbers),
	}
}

func (g *ValueGen) ValueForSetting(s seed.FieldTypeSetting) (any, error) {
	switch vt := s.(type) {
	case seed.StringSetting:
		return g.RangeStringLength(vt.MinCodePoints, vt.MaxCodePoints), nil
	case seed.BinarySetting:
		return RangeByteLength(g, vt.MinBytes, vt.MaxBytes), nil
	case seed.BooleanSetting:
		return Bool(g), nil
	case seed.TimeStampSetting:
		return g.RangeTime(vt.Min, vt.Max, vt.Scale, vt.WithTimeZoneOffset), nil
	case seed.IntegerSetting:
		if seed.Int64Setting().Covers(vt) {
			return g.RangeInt64(vt.Min.Int64(), vt.Max.Int64()), nil
		}
		return g.RangeBigInt(vt.Min, vt.Max), nil
	case seed.RealSetting:
		if vt.Standard == seed.Float64 {
			return g.RangeFloat64(*vt.MinFloat, *vt.MaxFloat), nil
		}
		return nil, seederrors.NewSystemError("ValueForSetting RealSetting.Standard %s not yet supported", vt.Standard)
	case seed.ReferenceSetting:
		return nil, seederrors.NewSystemError("ValueForSetting FieldTypeSetting ReferenceSetting not yet supported")
	case seed.ListSetting:
		return g.ValuesForSetting(vt.ItemTypeSetting, g.RangeInt64(vt.MinLength, vt.MaxLength))
	case seed.CombinationSetting:
		return g.MapForFieldGroup(&vt)
	}
	return nil, seederrors.NewSystemError("ValueForSetting FieldTypeSetting type=%T not handled", s)
}

func (g *ValueGen) MapForFieldGroup(fg seed.FieldGroupGetter) (map[seed.CodeName]any, error) {
	out := make(map[seed.CodeName]any, fg.GetFields().Count())
	err := fg.GetFields().RangeLogical(func(cn seed.CodeName, f *seed.Field) (err error) {
		out[cn], err = g.ValueForSetting(f.FieldTypeSetting)
		return
	})
	if err != nil {
		return nil, err
	}
	// make values more likely to be valid, does not have to be perfect
	err = seed.RangeRanges(fg, func(r seed.Range) error {
		start := out[r.Start]
		end := out[r.End]
		isLess, err2 := less(end, start)
		if err2 != nil {
			return err2
		}
		if isLess {
			out[r.Start], out[r.End] = end, start
		}
		return nil
	})
	return out, err
}

//nolint:forcetypeassert // use recover
func less(a, b any) (less bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = seederrors.CombineErrors(seederrors.NewSystemError(
				"recovered=%v", r), err)
		}
	}()
	switch at := a.(type) {
	case time.Time:
		return at.Before(b.(time.Time)), nil
	case int64:
		return at < b.(int64), nil
	}
	return false, seederrors.NewSystemError("comparison between %T and %T not implemented", a, b)
}

func (g *ValueGen) ValuesForSetting(s seed.FieldTypeSetting, length int64) ([]any, error) {
	out := make([]any, length)
	for i := range out {
		v, err := g.ValueForSetting(s)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

func (g *ValueGen) ValuesForObject(ob seed.ObjectGetter, length int) ([]map[seed.CodeName]any, error) {
	out := make([]map[seed.CodeName]any, length)
	for i := range out {
		v, err := g.MapForFieldGroup(ob)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}
