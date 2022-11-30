//nolint:forcetypeassert // using pre-generic library github.com/porfirion/trie
package dictionary

import (
	"github.com/porfirion/trie"
)

// prefixIndex wraps a trie implementation for possible future replacement.
type prefixIndex[V any] struct {
	internal *trie.Trie // this field should not be access outside this file
}

func makePrefixIndex[V any]() prefixIndex[V] {
	return prefixIndex[V]{
		internal: &trie.Trie{},
	}
}

/* unused
func (p prefixIndex[V]) count() int {
	return p.internal.Count()
}

func (p prefixIndex[V]) putCopyKey(k []byte, v V) {
	k = append(make([]byte, 0, len(k)), k...)
	p.putFast(k, v)
}
*/

func (p prefixIndex[V]) putFast(k []byte, v V) {
	p.internal.Put(k, v)
}

func (p prefixIndex[V]) getExact(key []byte) (V, bool) {
	value, ok := p.internal.Get(key)
	if !ok {
		var zeroValue V
		return zeroValue, false
	}
	return value.(V), true
}

func (p prefixIndex[V]) getAnyPrefixOf(longBytes []byte) (V, bool) {
	value, _, ok := p.internal.SearchPrefixIn(longBytes)
	if !ok {
		var zeroValue V
		return zeroValue, false
	}
	return value.(V), true
}

func (p prefixIndex[V]) getAnyInPrefix(prefix []byte) (V, bool) {
	sub, ok := p.internal.SubTrie(prefix, false) // false is faster code
	if !ok {
		var zeroValue V
		return zeroValue, false
	}
	var out trie.ValueType
	sub.Iterate(func(prefix []byte, value trie.ValueType) {
		out = value
	})
	return out.(V), true
}
