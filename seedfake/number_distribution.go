package seedfake

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"

	"github.com/xiegeo/must"
	"golang.org/x/exp/constraints"
)

type NumberDistribution interface {
	ByteDistribution
	IntegerDistribution
	RealDistribution
	// more could be added
}

type ByteDistribution interface {
	io.Reader
}

func Bool(d Int32Distribution) bool {
	pick := d.RangeInt32(0, 1)
	return pick == 1
}

func RangeByteLength(d NumberDistribution, min, max int64) []byte {
	length := d.RangeInt64(min, max)
	bs := make([]byte, length)
	_, _ = must.B2(d.Read(bs))(len(bs), nil)
	return bs
}

type Int32Distribution interface {
	RangeInt32(min, max int32) int32
}

type Int64Distribution interface {
	RangeInt64(min, max int64) int64
}

type IntegerDistribution interface {
	Int32Distribution
	Int64Distribution
	RangeBigInt(min, max *big.Int) *big.Int
	// more could be added
}

func RangeInt64[V ~int64 | ~int](d Int64Distribution, min, max V) V {
	return V(d.RangeInt64(int64(min), int64(max)))
}

type Float64Distribution interface {
	RangeFloat64(min, max float64) float64
}

type RealDistribution interface {
	Float64Distribution
	// more could be added
}

type Min struct{}

func (Min) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func (Min) RangeInt32(min, max int32) int32 {
	return min
}

func (Min) RangeInt64(min, max int64) int64 {
	return min
}

func (Min) RangeBigInt(min, max *big.Int) *big.Int {
	return new(big.Int).Set(min)
}

func (Min) RangeFloat64(min, max float64) float64 {
	return min
}

type Max struct{}

func (Max) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = math.MaxUint8
	}
	return len(p), nil
}

func (Max) RangeInt32(min, max int32) int32 {
	return max
}

func (Max) RangeInt64(min, max int64) int64 {
	return max
}

func (Max) RangeBigInt(min, max *big.Int) *big.Int {
	return new(big.Int).Set(max)
}

func (Max) RangeFloat64(min, max float64) float64 {
	return max
}

type Flat struct{ r rand.Rand }

func NewFlat(source rand.Source) *Flat {
	r := rand.New(source)
	return &Flat{*r}
}

func (f *Flat) Read(p []byte) (n int, err error) {
	return f.r.Read(p)
}

func (f *Flat) RangeInt32(min, max int32) int32 {
	diff := max - min + 1
	if diff < 0 { // If range is greater than Int31n can support.
		return int32(f.RangeInt64(int64(min), int64(max)))
	}
	return f.r.Int31n(diff) + min
}

func (f *Flat) RangeInt64(min, max int64) int64 {
	diff := max - min + 1
	if diff < 0 { // If range is greater than Int63n can support.
		return f.RangeBigInt(big.NewInt(min), big.NewInt(max)).Int64()
	}
	return f.r.Int63n(diff) + min
}

func (f *Flat) RangeBigInt(min, max *big.Int) *big.Int {
	temp := big.NewInt(1)
	temp.Add(temp, max) // One greater than max to include max value in output.
	temp.Sub(temp, min)
	rint := must.V(crand.Int(&f.r, temp)) // random source f.r should never return error
	return temp.Add(rint, min)
}

// RangeFloat64
// rewrite Rand.Float64 to include max,
// but still not perfect for lest significant bits
func (f *Flat) RangeFloat64(min, max float64) float64 {
	if math.IsInf(min, -1) {
		min = -math.MaxFloat64
	}
	if math.IsInf(max, 1) {
		max = math.MaxFloat64
	}
	diff := max - min
	if diff > math.MaxFloat64 {
		must.True(min < 0 && max > 0, "if diff is greater than max float, range must have crossed zero")
		switch choseWeighted(&f.r, -min, max) {
		case 0:
			return -f.RangeFloat64(0, -min)
		case 1:
			return f.RangeFloat64(0, max)
		}
	}
	return float64(f.r.Int63())/(1<<63)*(max-min) + min //nolint:gomnd // magic number 63 from Int63
}

type Mixed struct {
	pickBy     *rand.Rand
	dists      []NumberDistribution
	pickLevels []float64
}

func NewMixedDistribution(pickBy *rand.Rand, dists []NumberDistribution, weights []float64) *Mixed {
	if len(dists) != len(weights) {
		panic("NewMixedDistribution: length not same")
	}
	sum, weights := sumNoOverflow(weights...)
	pickLevels := make([]float64, len(weights)-1)
	var current float64
	for i, w := range weights[:len(weights)-1] {
		current += w / sum
		pickLevels[i] = current
	}
	return &Mixed{
		pickBy:     pickBy,
		dists:      dists,
		pickLevels: pickLevels,
	}
}

func NewMinMaxFlat(s rand.Source, min, max, flat float64) *Mixed {
	return NewMixedDistribution(rand.New(s),
		[]NumberDistribution{Min{}, Max{}, NewFlat(s)},
		[]float64{min, max, flat})
}

func (m *Mixed) pickDistribution() NumberDistribution {
	pick := m.pickBy.Float64()
	for i, l := range m.pickLevels {
		if l > pick {
			return m.dists[i]
		}
	}
	return m.dists[len(m.dists)-1]
}

func (m *Mixed) Read(p []byte) (n int, err error) {
	return m.pickDistribution().Read(p)
}

func (m *Mixed) RangeInt32(min, max int32) int32 {
	return m.pickDistribution().RangeInt32(min, max)
}

func (m *Mixed) RangeInt64(min, max int64) int64 {
	return m.pickDistribution().RangeInt64(min, max)
}

func (m *Mixed) RangeBigInt(min, max *big.Int) *big.Int {
	return m.pickDistribution().RangeBigInt(min, max)
}

func (m *Mixed) RangeFloat64(min, max float64) float64 {
	return m.pickDistribution().RangeFloat64(min, max)
}

func mustNotNeg[V constraints.Signed | constraints.Float](v V) {
	if v < 0 {
		panic(fmt.Sprintf("negative value %v of %T", v, v))
	}
}

// sumNoOverflow returns the sum of ss. If sum of ss overflows to infinity, ss is scaled down
// so that sum does not overflow.
func sumNoOverflow(ss ...float64) (float64, []float64) {
	var largest float64
	for _, s := range ss {
		mustNotNeg(s)
		if s > largest {
			largest = s
			must.False(math.IsInf(s, 0), "infinity will always overflow")
		}
	}
	if math.MaxFloat64/float64(len(ss)) <= largest {
		for i, s := range ss {
			ss[i] = s / (1 << 31) //nolint:gomnd // Some fast deviser that make sure sum will not overflow.
		}
	}
	var sum float64
	for _, s := range ss {
		sum += s
	}
	return sum, ss
}

func choseWeighted(r *rand.Rand, weights ...float64) int {
	sum, weights := sumNoOverflow(weights...)
	pick := r.Float64() * sum
	var current float64
	for i, w := range weights[:len(weights)-1] {
		current += w
		if current > pick {
			return i
		}
	}
	return len(weights)
}

func pickFromSlice[V any](dist Int32Distribution, vs []V) V {
	pick := dist.RangeInt32(0, int32(len(vs)-1))
	return vs[pick]
}
