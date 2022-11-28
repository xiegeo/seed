package dictionary

import (
	"github.com/xiegeo/seed/seederrors"
	"golang.org/x/exp/constraints"
)

// Dictionary of seed.CodeName (use generics to avoid import cycle) to a field or object.
// Dictionary enforces seed name rules on character set, prefix, and versioning.
type Dictionary[K ~string, V any] struct {
	m           map[K]V
	prefixIndex prefixIndex[[]K]
}

func NewDictionary[K ~string, V any]() *Dictionary[K, V] {
	return &Dictionary[K, V]{
		m:           make(map[K]V),
		prefixIndex: makePrefixIndex[[]K](),
	}
}

func (d *Dictionary[K, V]) set(k K, v V, simple []byte, version int8) error {
	d.m[k] = v
	byVersion, _ := d.prefixIndex.getExact(simple)
	setSliceValue(version, byVersion, k)
	d.prefixIndex.putFast(simple, byVersion)
	return nil
}

func (d *Dictionary[K, V]) Add(k K, v V) error {
	simple, version, err := Simplify(k)
	if err != nil {
		return err
	}
	if version < 1 {
		if version != -1 {
			return seederrors.NewSystemError("unexpected version from Simplify %d", version)
		}
		version = 0 // convert -1 to 0 for indexing
	}
	exactList, found := d.prefixIndex.getExact(simple)
	if found {
		exactVersion := getSliceValue(version, exactList)
		if len(exactVersion) == 0 {
			// found name part and not version part, this is a different version of the field.
			return d.set(k, v, simple, version)
		}
		return seederrors.NewNameVersionRepeatedError(k, exactVersion, version)
	}
	// do prefix checks without version postfix
	longerName, found := d.prefixIndex.getAnyInPrefix(simple)
	if found {
		return seederrors.NewNameRepeatedError(k, longerName[0])
	}
	shorterName, found := d.prefixIndex.getAnyPrefixOf(simple)
	if found {
		return seederrors.NewNameRepeatedError(shorterName[0], k)
	}
	return d.set(k, v, simple, version)
}

func getSliceValue[I constraints.Integer, V any](i I, s []V) V {
	if int(i) <= len(s) {
		var zeroValue V
		return zeroValue
	}
	return s[i]
}

func setSliceValue[I constraints.Integer, V any](i I, s []V, v V) []V {
	if int(i) <= len(s) {
		s = append(s, make([]V, int(i)-len(s)+1)...)
	}
	s[i] = v
	return s
}
