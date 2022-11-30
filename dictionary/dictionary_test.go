package dictionary

import (
	"fmt"
	"regexp"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDictonary(t *testing.T) {
	testCases := []struct {
		name      string
		fits      []string
		notFit    []string
		objFits   []string
		objNotFit []string
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
			objFits: []string{
				"AA", "BBC", "CCCP",
				"vvv2",
			},
			objNotFit: []string{
				"av2", "A_v3", "aV_7", // exact match
			},
		},
	}
	for _, tC := range testCases {
		// t.Run(tC.name, func(t *testing.T) { // only one case for now
		t.Run("one to one match", func(t *testing.T) {
			for i := 0; i < len(tC.notFit); i++ {
				a := tC.fits[i]
				b := tC.notFit[i]
				t.Run(fmt.Sprint(a, " vs ", b), func(t *testing.T) {
					dict := NewField[string, int]()
					require.NoError(t, dict.Add(a, i), a)
					require.Error(t, dict.Add(b, i), b)
					// reverted
					dict = NewField[string, int]()
					require.NoError(t, dict.Add(b, i), b)
					require.Error(t, dict.Add(a, i), a)
				})
			}
		})
		allFits := NewField[string, int]()
		t.Run("insert all", func(t *testing.T) {
			for i, k := range tC.fits {
				assert.NoError(t, allFits.Add(k, i), k)
			}
		})
		allRevese := NewField[string, int]()
		t.Run("insert all in reverse", func(t *testing.T) {
			for i, k := range reverseAppend(nil, tC.fits...) {
				assert.NoError(t, allRevese.Add(k, len(tC.fits)-i-1), k)
			}
			allFits2 := *allFits
			allFits2.logicalOrder = reverseAppend(nil, allFits.logicalOrder...)
			assert.Equal(t, &allFits2, allRevese)
		})
		t.Run("not fit", func(t *testing.T) {
			for i, k := range tC.notFit {
				assert.Error(t, allFits.Add(k, i), k)
			}
		})
		allFits = NewField[string, int]()
		t.Run("reverse not fit", func(t *testing.T) {
			for i, k := range tC.notFit {
				allFits.Add(k, i)
			}
			for i, k := range tC.fits {
				assert.Error(t, allFits.Add(k, i), k)
			}
		})
		allFits = NewObject[string, int]()
		t.Run("allowPrefixMatch", func(t *testing.T) {
			for i, k := range tC.fits {
				assert.NoError(t, allFits.Add(k, i), k)
			}
			for i, k := range tC.objFits {
				assert.NoError(t, allFits.Add(k, i), k)
			}
			for i, k := range tC.objNotFit {
				assert.Error(t, allFits.Add(k, i), k)
			}
			for i, k := range tC.fits {
				assert.Error(t, allFits.Add(k, i), k)
			}
			for i, k := range tC.objFits {
				assert.Error(t, allFits.Add(k, i), k)
			}
		})
		// })
	}
}

func reverseAppend[T any](a []T, b ...T) []T {
	for i := len(b) - 1; i >= 0; i-- {
		a = append(a, b[i])
	}
	return a
}

func FuzzDictionary(f *testing.F) {
	allowed := regexp.MustCompile("^[_a-zA-Z0-9]+$")
	f.Add("VV2", uint16(1), false)  // v.v2 not allowed because version can not start a field.
	f.Add("aV2a", uint16(3), false) // av2.a allowed but av2a is not allow as a field.
	f.Add("Aİ0", uint16(4), false)  // regression for upper case of i is not just I.
	f.Add("a0V2", uint16(3), false) // regression for bug building error message.
	f.Fuzz(func(t *testing.T, b string, aLenth uint16, allowPrefix bool) {
		if len(b) > 10 || int(aLenth) > len(b) {
			return // discourage long or uninteresting cases
		}
		// make such a+c = b
		a := b[:aLenth]
		c := b[aLenth:]
		dict := NewField[string, string]()
		dict.allowPrefixMatch = allowPrefix
		dict2 := NewField[string, string]()
		dict2.allowPrefixMatch = allowPrefix
		if err := dict.Add(a, a); err != nil {
			_, _, err2 := Simplify(a) // adding to a fresh dictionary can only produce name simplification errors.
			require.EqualError(t, err, err2.Error())
			return
		}
		if !utf8.ValidString(a) {
			t.Fatal("invalid string was added to dictionary")
		}
		if !allowed.MatchString(a) {
			t.Fatal("illegal character was added to dictionary")
		}
		errBafterA := dict.Add(b, b)
		errBfirst := dict2.Add(b, b)
		if errBfirst != nil {
			// If dictionary is empty, then the only error can be b is a illegal name.
			// If b is a illegal name, then it must be the error reported on Add.
			require.EqualError(t, errBfirst, errBafterA.Error())
		} else {
			// if b is legal and ...
			errAafterB := dict2.Add(a, a)
			require.Equal(t, dict.Count(), dict2.Count())
			if errBafterA != nil {
				// if b can not be added after a, then a can not be added after b
				require.EqualError(t, errAafterB, errBafterA.Error())
			} else {
				// if b can be added after a, then a can be added after b
				require.NoError(t, errAafterB)
				require.Equal(t, dict.m, dict2.m)
				require.NotEqual(t, dict.logicalOrder, dict2.logicalOrder)
				require.Equal(t, dict.prefixIndex, dict2.prefixIndex)
			}
		}
		if errBafterA != nil {
			return
		}

		// known a+c = b, and a, b can be in the same dictionary,
		// then c can not be legal, if allowPrefix is false.
		if !allowPrefix {
			_, _, err := Simplify(c)
			require.Error(t, err)
		}
	})
}
