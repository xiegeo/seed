package seed

import (
	"sync"

	locale "github.com/jeandeaual/go-locale"
	"github.com/puzpuzpuz/xsync/v2"
	"golang.org/x/text/language"
)

// I18n[T any] is used for internationalization, such as text and icons displayed to users.
type I18n[T any] map[language.Tag]T

var _ I18nGetter[any] = I18n[any]{}

type I18nGetter[T any] interface {
	GetValue(p *Picker, fallback T) T
	Count() int
	RangeAll(func(language.Tag, T))
}

func NewI18n[T any](from I18nGetter[T]) I18n[T] {
	out := make(I18n[T], from.Count())
	from.RangeAll(func(k language.Tag, v T) {
		out[k] = v
	})
	return out
}

func (n I18n[T]) Count() int {
	return len(n)
}

func (n I18n[T]) RangeAll(f func(language.Tag, T)) {
	for k, v := range n {
		f(k, v)
	}
}

func (n I18n[T]) GetValue(p *Picker, fallback T) T {
	lang, _ := Pick[T](p, n)
	v, ok := n[lang]
	if ok {
		return v
	}
	return fallback
}

// Picker is an inverse language.Matcher, where user request language is static
// but supported language is dynamic.
//
// Picker is safe to use concurrently.
type Picker struct {
	preferred        []language.Tag
	comprehendsCache *xsync.MapOf[string, pickOrder]
	fallback         *Picker
}

// NewPicker creates a language picker by confidence, then by natural order
// of the preferred list. If no matches are found, fallback is used.
func NewPicker(preferred []language.Tag, fallback *Picker) *Picker {
	return &Picker{
		preferred:        preferred,
		comprehendsCache: xsync.NewMapOf[pickOrder](),
		fallback:         fallback,
	}
}

var (
	_systemLogPicker           *Picker
	_systemLogPickerOnce       sync.Once
	_systemLogPickerInitErrors []error
)

// SystemLogPicker picks the language based on the configuration of the local environment.
func SystemLogPicker() *Picker {
	_systemLogPickerOnce.Do(func() {
		// get the system locales from the current user environment.
		systemLocales, err := locale.GetLocales()
		if err != nil {
			_systemLogPickerInitErrors = append(_systemLogPickerInitErrors, err)
		}
		// convert strings to tags
		userTags := make([]language.Tag, 0, len(systemLocales))
		for _, v := range systemLocales {
			tag, err := language.Parse(v)
			if err != nil {
				_systemLogPickerInitErrors = append(_systemLogPickerInitErrors, err)
				continue
			}
			userTags = append(userTags, tag)
		}
		_systemLogPicker = NewPicker(userTags, _lastPicker)
	})
	return _systemLogPicker
}

var _lastPicker *Picker // picker of last resort, used as the last fallback picker

func init() {
	_lastPicker = NewPicker([]language.Tag{language.English}, nil)
}

type pickOrder struct {
	confidence language.Confidence
	order      int
}

func (p pickOrder) betterThen(p2 pickOrder) bool {
	if p.confidence == p2.confidence {
		return p.order < p2.order
	}
	return p.confidence > p2.confidence
}

// Pick picks the best value from I18n by the picker's preference.
// If no languages matched, then (und, false) is returned.
func Pick[T any](picker *Picker, values I18nGetter[T]) (language.Tag, bool) {
	var bestTag language.Tag
	var bestOrder pickOrder
	values.RangeAll(func(tag language.Tag, v T) {
		order := picker.getOrder(tag)
		if order.betterThen(bestOrder) {
			bestTag = tag
			bestOrder = order
		}
	})
	if picker.fallback != nil && bestOrder.confidence == language.No {
		return Pick(picker.fallback, values)
	}
	return bestTag, bestOrder.confidence > 0
}

func (p *Picker) getOrder(alternative language.Tag) pickOrder {
	aString := alternative.String()
	order, ok := p.comprehendsCache.Load(aString)
	if ok {
		return order
	}
	var bestOrder pickOrder
	for i, speeker := range p.preferred {
		c := language.Comprehends(speeker, alternative)
		newOrder := pickOrder{confidence: c, order: i}
		if newOrder.betterThen(bestOrder) {
			bestOrder = newOrder
		}
	}
	p.comprehendsCache.Store(aString, bestOrder)
	return bestOrder
}
