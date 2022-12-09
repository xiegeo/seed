package dictionary

// SelfKeyed is a dictionary keyed by a keying function that derives key from value.
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

// New creates a new empty SelfKeyed dictionary with the same type and configuration.
func (d *SelfKeyed[K, V]) New() *SelfKeyed[K, V] {
	return &SelfKeyed[K, V]{
		key:        d.key,
		Dictionary: *d.Dictionary.New(),
	}
}

// NewMap creates a new SelfKeyed dictionary with values mapped from the old one,
// keys are updated by the keying function.
func (d *SelfKeyed[K, V]) NewMap(f func(V) (V, error)) (*SelfKeyed[K, V], error) {
	dict := d.New()
	err := d.RangeLogical(func(k K, v V) error {
		newValue, err := f(v)
		if err != nil {
			return err
		}
		return dict.AddValue(newValue)
	})
	if err != nil {
		return nil, err
	}
	return dict, nil
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
