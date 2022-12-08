package seedfake

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlat(t *testing.T) {
	source := rand.NewSource(0)
	flat := NewFlat(source)

	testBool(t, flat, 1000, 469)

	testInt64(t, flat, -10, 10, 100, big.NewInt(42), big.NewInt(506))
	sum, _ := big.NewInt(0).SetString("57982752274955655778", 10) // near 5x MaxInt64
	testInt64(t, flat, 0, math.MaxInt64, 10, sum, sum)
	sum, _ = big.NewInt(0).SetString("40192256412768239738", 10)
	testInt64(t, flat, 1, math.MaxInt64, 10, sum, sum) // take a different code path
	sum, _ = big.NewInt(0).SetString("-12498007953715184557", 10)
	sumAbs, _ := big.NewInt(0).SetString("58550210712764616929", 10)
	testInt64(t, flat, -math.MaxInt64, math.MaxInt64, 10, sum, sumAbs) // full range
}

func testBool(t *testing.T, d Int32Distribution, repeats, expects int) {
	t.Run("Bool", func(t *testing.T) {
		var got int
		for i := repeats; i > 0; i-- {
			if Bool(d) {
				got++
			}
		}
		assert.Equal(t, expects, got)
	})
}

func testInt64(t *testing.T, d Int64Distribution, min, max, repeats int64, expSum *big.Int, expAbsSum *big.Int) {
	t.Run(fmt.Sprint("Int64 ", min, max), func(t *testing.T) {
		got := big.NewInt(0)
		gotAbs := big.NewInt(0)
		for i := repeats; i > 0; i-- {
			n := big.NewInt(d.RangeInt64(min, max))
			got.Add(got, n)
			gotAbs.Add(gotAbs, n.Abs(n))
		}
		assert.Equal(t, expSum, got)
		assert.Equal(t, expAbsSum, gotAbs)
	})
}
