package seedfake

import (
	"bytes"
	"unicode"
)

type StringDistribution interface {
	RangeStringLength(min, max int64) string
}

type Rune struct {
	dist  IntegerDistribution
	chars [][2]rune
	buf   bytes.Buffer
}

func NewRune(dist IntegerDistribution, charRanges ...[2]rune) *Rune {
	return &Rune{
		dist:  dist,
		chars: charRanges,
	}
}

func (r *Rune) Rune() rune {
	if len(r.chars) == 0 {
		return r.dist.RangeInt32(0, unicode.MaxRune)
	}
	charRange := pickFromSlice(r.dist, r.chars)
	return r.dist.RangeInt32(charRange[0], charRange[1])
}

func (r *Rune) RangeStringLength(min, max int64) string {
	length := r.dist.RangeInt64(min, max)
	for i := 0; i < int(length); i++ {
		r.buf.WriteRune(r.Rune())
	}
	out := r.buf.String()
	r.buf.Reset()
	return out
}
