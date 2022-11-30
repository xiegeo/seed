package dictionary

type SelfKeyed[K ~string, V any] struct {
	key func(V) K
	Dictionary[K, V]
}

// NewSelfKeyed create a SelfKeyed dictionary that have conveniences for add by values that knows their
// own key. The passed in dictionary must not be used directly anymore.
func NewSelfKeyed[K ~string, V any](d *Dictionary[K, V], key func(V) K) *SelfKeyed[K, V] {
	d0 := *d
	d.m = nil // make the passed in dictionary unusable
	return &SelfKeyed[K, V]{
		key:        key,
		Dictionary: d0,
	}
}

func (d *SelfKeyed[K, V]) AddValue(vs ...V) error {
	for _, v := range vs {
		k := d.key(v)
		err := d.Add(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
