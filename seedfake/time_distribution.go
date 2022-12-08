package seedfake

import (
	"fmt"
	"math"
	"math/big"
	"time"
)

type TimeDistribution interface {
	RangeTime(min, max time.Time, scale time.Duration) time.Time
}

type Time struct {
	dist IntegerDistribution
}

func NewTime(dist IntegerDistribution) *Time {
	return &Time{dist: dist}
}

// Zone returns a time.Location from distribution.
//
// Range is taken by the minimum of following:
//   - Prior to MySQL 8.0.19, this value had to be in the range '-12:59' to '+13:00', inclusive;
//     beginning with MySQL 8.0.19, the permitted range is '-13:59' to '+14:00', inclusive.
//   - PostgreSQL(15) time [ (p) ] with time zone; low value: 00:00:00+1559 high value: 24:00:00-1559
//
//nolint:gomnd
func (t *Time) Zone() *time.Location {
	offsetMinutes := RangeInt64(t.dist, -13*60+1, +13*60)
	return time.FixedZone(fmt.Sprintf("%+02d:%02d", offsetMinutes/60, abs(offsetMinutes%60)), offsetMinutes*60)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (t *Time) RangeTime(min, max time.Time, scale time.Duration) time.Time {
	diff := max.Sub(min)
	if diff != math.MaxInt64 {
		addition := t.dist.RangeInt64(0, int64(diff/scale)) * int64(scale)
		return min.Add(time.Duration(addition)).In(t.Zone())
	}
	bigIntMin := timeToBigInt(min, scale)
	target := t.dist.RangeBigInt(bigIntMin, timeToBigInt(max, scale))
	return bigIntToTime(target, scale).In(t.Zone())
}

func timeToBigInt(t time.Time, scale time.Duration) *big.Int {
	micro := big.NewInt(t.UnixMicro())
	if scale%time.Microsecond == 0 {
		return micro
	}
	return micro.Mul(micro, big.NewInt(int64(time.Microsecond)))
}

func bigIntToTime(b *big.Int, scale time.Duration) time.Time {
	if scale%time.Microsecond == 0 {
		return time.UnixMicro(b.Int64())
	}
	high, low := big.NewInt(0).DivMod(b, big.NewInt(int64(time.Microsecond)), big.NewInt(0))
	return time.UnixMicro(high.Int64()).Add(time.Duration(low.Int64()))
}
