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
					dict := New[string, int]()
					require.NoError(t, dict.Add(a, i), a)
					require.Error(t, dict.Add(b, i), b)
					// reverted
					dict = New[string, int]()
					require.NoError(t, dict.Add(b, i), b)
					require.Error(t, dict.Add(a, i), a)
				})
			}
		})
		allFits := New[string, int]()
		t.Run("insert all", func(t *testing.T) {
			for i, k := range tC.fits {
				assert.NoError(t, allFits.Add(k, i), k)
			}
		})
		allRevese := New[string, int]()
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
		allFits = New[string, int]()
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

func FuzzDictionary(f *testing.F) {
	allowed := regexp.MustCompile("^[_a-zA-Z0-9]+$")
	f.Add("VV2", uint16(1))  // v.v2 not allowed because version can not start a field
	f.Add("aV2a", uint16(3)) // av2.a allowed but av2a is not allow as a field
	f.Fuzz(func(t *testing.T, b string, aLenth uint16) {
		if len(b) > 10 || int(aLenth) > len(b) {
			return // discourage long or uninteresting cases
		}
		// make such a+c = b
		a := string(b[:aLenth])
		c := string(b[aLenth:])
		dict := New[string, string]()
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
		errB := dict.Add(b, b)

		dict2 := New[string, string]()
		errB2 := dict2.Add(b, b)
		if errB2 != nil {
			// If dictionary is empty, then the only error can be b is a illegal name.
			// If b is a illegal name, then it must be the error reported on Add.
			require.EqualError(t, errB2, errB.Error())
		} else {
			// if b is legal and ...
			err := dict2.Add(a, a)
			if errB != nil {
				// if b can not be added after a, then a can not be added after b
				require.EqualError(t, err, errB.Error())
			} else {
				// if b can be added after a, then a can be added after b
				require.NoError(t, err)
			}
		}

		// if a+c = b, and
		// if a, b can be in the same dictionary,
		// then c can not be legal.
		if errB != nil {
			return
		}
		_, _, err := Simplify(c)
		require.Error(t, err)
	})
}
