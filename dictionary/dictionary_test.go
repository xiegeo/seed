package dictionary

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDictonary(t *testing.T) {
	testCases := []struct {
		name   string
		fits   []string
		notFit []string
	}{
		{
			name: "",
			fits: []string{
				"a", "bb", "ccc",
				"av2", "av3", "av7",
				"v", "vv2",
			},
			notFit: []string{
				"AA", "BBC", "CCCP", // got prefixed
				"av2", "A_v3", "aV_7", // exact match
				"vvv2", "vvv2", // got prefixed even with version removed
			},
		},
	}
	for _, tC := range testCases {
		// t.Run(tC.name, func(t *testing.T) {
		t.Run("one to one match", func(t *testing.T) {
			for i := 0; i < len(tC.notFit); i++ {
				a := tC.fits[i]
				b := tC.notFit[i]
				t.Run(fmt.Sprint(a, " vs ", b), func(t *testing.T) {
					dict := NewDictionary[string, int]()
					require.NoError(t, dict.Add(a, i), a)
					require.Error(t, dict.Add(b, i), b)
					// reverted
					dict = NewDictionary[string, int]()
					require.NoError(t, dict.Add(b, i), b)
					require.Error(t, dict.Add(a, i), a)
				})
			}
		})
		allFits := NewDictionary[string, int]()
		t.Run("insert all", func(t *testing.T) {
			for i, k := range tC.fits {
				assert.NoError(t, allFits.Add(k, i), k)
			}
		})
		allRevese := NewDictionary[string, int]()
		t.Run("insert all in reverse", func(t *testing.T) {
			for i, k := range reverseAppend(nil, tC.fits...) {
				assert.NoError(t, allRevese.Add(k, len(tC.fits)-i-1), k)
			}
			assert.Equal(t, allFits, allRevese)
		})
		t.Run("not fit", func(t *testing.T) {
			for i, k := range tC.notFit {
				assert.Error(t, allFits.Add(k, i), k)
			}
		})
		allFits = NewDictionary[string, int]()
		t.Run("reverse not fit", func(t *testing.T) {
			for i, k := range tC.notFit {
				allFits.Add(k, i)
			}
			for i, k := range tC.fits {
				assert.Error(t, allFits.Add(k, i), k)
			}
		})
		//})
	}
}

func reverseAppend[T any](a []T, b ...T) []T {
	for i := len(b) - 1; i >= 0; i-- {
		a = append(a, b[i])
	}
	return a
}
