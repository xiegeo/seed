package seed

import (
	"hash/maphash"
	"sync"

	locale "github.com/jeandeaual/go-locale"
	"github.com/puzpuzpuz/xsync/v2"
	"golang.org/x/text/language"
)

// I18n[T any] is used for internationalization, such as text and icons displayed to users.
type I18n[T any] map[language.Tag]T

func (n I18n[T]) GetValue(p *Picker, fallback T) T {
	lang, _ := Pick(p, n)
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
	comprehendsCache *xsync.MapOf[language.Tag, pickOrder]
	fallback         *Picker
}

// NewPicker creates a language picker by confidence, then by natural order
// of the preferred list. If no matches are found, fallback is used.
func NewPicker(preferred []language.Tag, fallback *Picker) *Picker {
	return &Picker{
		preferred: preferred,
		comprehendsCache: xsync.NewTypedMapOf[language.Tag, pickOrder](func(s maphash.Seed, t language.Tag) uint64 {
			var h maphash.Hash
			h.SetSeed(s)
			text, err := t.MarshalText()
			if err != nil {
				panic(err) // MarshalText should not error
			}
			h.Write(text)
			return h.Sum64()
		}),
		fallback: fallback,
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
	tag        language.Tag
}

func (p pickOrder) betterThen(p2 pickOrder) bool {
	if p.confidence == p2.confidence {
		return p.order < p2.order
	}
	return p.confidence > p2.confidence
}

// Pick picks the best value from I18n by the picker's preference.
// If no languages matched, then (und, false) is returned.
func Pick[T any](p *Picker, values I18n[T]) (language.Tag, bool) {
	var bestTag language.Tag
	var bestOrder pickOrder
	for tag := range values {
		order := p.getOrder(tag)
		if order.betterThen(bestOrder) {
			bestTag = tag
			bestOrder = order
		}
	}
	if p.fallback != nil && bestOrder.confidence == language.No {
		return Pick(p.fallback, values)
	}
	return bestTag, bestOrder.confidence > 0
}

func (p *Picker) getOrder(alternative language.Tag) pickOrder {
	order, ok := p.comprehendsCache.Load(alternative)
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
	p.comprehendsCache.Store(alternative, bestOrder)
	return bestOrder
}
