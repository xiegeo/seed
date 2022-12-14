package dictionary

var (
	// Dictionary and SelfKeyed implements getter
	_ Getter[string, any] = &Dictionary[string, any]{}
	_ Getter[string, any] = &SelfKeyed[string, any]{}
)

type Getter[K ~string, V any] interface {
	private() // allow interface to be extendable
	RangeLogical(f func(K, V) error) error
	Count() int
	Get(k K) (V, bool)
	Values() []V
}

// convertor
type convertor[K ~string, V, V2 any] struct {
	Getter[K, V]
	conv func(V) V2
}

func (c convertor[K, V, V2]) RangeLogical(f func(K, V2) error) error {
	return c.Getter.RangeLogical(func(k K, v V) error {
		return f(k, c.conv(v))
	})
}

func (c convertor[K, V, V2]) Get(k K) (V2, bool) {
	v, ok := c.Getter.Get(k)
	return c.conv(v), ok
}

func (c convertor[K, V, V2]) Values() []V2 {
	return Values(c.RangeLogical, c.Count())
}

func MapValue[K ~string, V, V2 any](from Getter[K, V], convFunc func(V) V2) Getter[K, V2] {
	return convertor[K, V, V2]{Getter: from, conv: convFunc}
}
