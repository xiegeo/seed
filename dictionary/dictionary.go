package dictionary

import (
	"golang.org/x/exp/constraints"

	"github.com/xiegeo/seed/seederrors"
)

// Dictionary of seed.CodeName (use generics to avoid import cycle) to a field or object definition.
// Dictionary enforce naming rules on character-set, and versioning.
// Dictionary preserves assertion/logical ordering, useful for field names.
//
// Dictionary optionally enforces field name rules on prefix.
//   - When used for fields, AllowPrefixMatch is set to false (use NewField).
//   - When used for objects, AllowPrefixMatch is set to true (use NewObject).
//
// Duplication is checked on simplified versions of code names by removing case and "_".
// "AB", "aB", "A_b" all simplify to "ab". A field can not have a name that is the prefix
// of another under the same parent (except "v2", "v3"... postfixes). All fields of the
// same name must have the same properties (except label and description).
//
// For details on simplification, see func Simplify.
//
// Noticeable none features: delete or replace k-v pairs. Instead, domains are expected to be imported,
// and could just be reimported when upstream changes.
// So, dynamically changing a domain do not have a use case yet.
// While data migration to support domain modification is another beast all together.
type Dictionary[K ~string, V any] struct {
	m                map[K]V
	logicalOrder     []K
	prefixIndex      prefixIndex[[]K] // simplified name -> version number -> full name
	allowPrefixMatch bool
}

// NewField creates a dictionary for a list of fields in an object
func NewField[K ~string, V any]() *Dictionary[K, V] {
	return &Dictionary[K, V]{
		m:           make(map[K]V),
		prefixIndex: makePrefixIndex[[]K](),
	}
}

// NewObject creates a dictionary for a list of objects in a domain
func NewObject[K ~string, V any]() *Dictionary[K, V] {
	dict := NewField[K, V]()
	dict.allowPrefixMatch = true
	return dict
}

// New creates a new empty dictionary with the same type and configuration.
func (d *Dictionary[K, V]) New() *Dictionary[K, V] {
	dict := NewField[K, V]()
	dict.allowPrefixMatch = d.allowPrefixMatch
	return dict
}

// RangeLogical range the dictionary in logical (insertion) order, stops on first error encountered.
func (d *Dictionary[K, V]) RangeLogical(f func(K, V) error) error {
	for _, k := range d.logicalOrder {
		v, ok := d.m[k]
		if !ok {
			return seederrors.NewSystemError("Dictionary internals is inconsistent: logicalOrder key %s not found in map", k)
		}
		err := f(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dictionary[K, V]) Count() int {
	return len(d.m)
}

func (d *Dictionary[K, V]) Get(k K) (V, bool) {
	v, ok := d.m[k]
	return v, ok
}

func (d *Dictionary[K, V]) set(k K, v V, simple []byte, version int8) error {
	_, exist := d.m[k]
	if !exist {
		d.logicalOrder = append(d.logicalOrder, k)
	}
	d.m[k] = v
	byVersion, _ := d.prefixIndex.getExact(simple)
	byVersion = setSliceValue(version, byVersion, k)
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
	if d.allowPrefixMatch {
		return d.set(k, v, simple, version)
	}
	// do prefix checks without version postfix
	longerName, found := d.prefixIndex.getAnyInPrefix(simple)
	if found {
		return seederrors.NewNameRepeatedError(k, getLastValue(longerName))
	}
	shorterName, found := d.prefixIndex.getAnyPrefixOf(simple)
	if found {
		return seederrors.NewNameRepeatedError(getLastValue(shorterName), k)
	}
	return d.set(k, v, simple, version)
}

func getSliceValue[I constraints.Integer, V any](i I, s []V) V {
	if int(i) >= len(s) {
		var zeroValue V
		return zeroValue
	}
	return s[i]
}

func setSliceValue[I constraints.Integer, V any](i I, s []V, v V) []V {
	if int(i) >= len(s) {
		s = append(s, make([]V, int(i)-len(s)+1)...)
	}
	s[i] = v
	return s
}

func getLastValue[V any](s []V) V {
	if len(s) == 0 {
		var zeroValue V
		return zeroValue
	}
	return s[len(s)-1]
}
