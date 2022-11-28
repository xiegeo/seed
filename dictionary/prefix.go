//nolint:forcetypeassert // using pre-generic library github.com/porfirion/trie
package dictionary

import (
	"github.com/porfirion/trie"
)

// prefixIndex wraps a trie implementation for possible future replacement.
type prefixIndex[V any] struct {
	interal *trie.Trie // this field should not be access outside this file
}

func makePrefixIndex[V any]() prefixIndex[V] {
	return prefixIndex[V]{
		interal: &trie.Trie{},
	}
}

func (p prefixIndex[V]) count() int {
	return p.interal.Count()
}

func (p prefixIndex[V]) putFast(k []byte, v V) {
	p.interal.Put(k, v)
}

func (p prefixIndex[V]) putCopyKey(k []byte, v V) {
	k = append(make([]byte, 0, len(k)), k...)
	p.putFast(k, v)
}

func (p prefixIndex[V]) delete(keys ...[]byte) int {
	filterOut := makePrefixIndex[V]()
	var zeroValue V
	for _, k := range keys {
		filterOut.putFast(k, zeroValue)
	}
	return p.deleteIndex(filterOut)
}

func (p prefixIndex[V]) deleteIndex(filterOut prefixIndex[V]) int {
	newIndex := makePrefixIndex[V]()
	var count int
	p.interal.Iterate(func(key []byte, value trie.ValueType) {
		if _, found := filterOut.interal.Get(key); found {
			count++
			return
		}
		newIndex.putCopyKey(key, value.(V))
	})
	*p.interal = *newIndex.interal
	return count
}

func (p prefixIndex[V]) getExact(key []byte) (V, bool) {
	value, ok := p.interal.Get(key)
	if !ok {
		var zeroValue V
		return zeroValue, false
	}
	return value.(V), true
}

func (p prefixIndex[V]) getAnyPrefixOf(longBytes []byte) (V, bool) {
	value, _, ok := p.interal.SearchPrefixIn(longBytes)
	if !ok {
		var zeroValue V
		return zeroValue, false
	}
	return value.(V), true
}

func (p prefixIndex[V]) getAnyInPrefix(prefix []byte) (V, bool) {
	sub, ok := p.interal.SubTrie(prefix, false) // false is faster code
	if !ok {
		var zeroValue V
		return zeroValue, false
	}
	vs := sub.GetAll(nil)
	return vs[0].(V), true
}
