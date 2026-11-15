package sniffer

import "fmt"

// Plugins is filled once at startup and never changed afterwards,
// so it needs no locks.
type Plugins struct {
	locators   []Locator
	extractors map[string]Extractor // keyed by Target.Kind
}

func NewPlugins() *Plugins {
	return &Plugins{extractors: make(map[string]Extractor)}
}

func (p *Plugins) AddLocator(l Locator) {
	p.locators = append(p.locators, l)
}

func (p *Plugins) AddExtractor(e Extractor) {
	for _, kind := range e.Kinds() {
		if _, dup := p.extractors[kind]; dup {
			panic(fmt.Sprintf("sniffer: two extractors registered for kind %q", kind))
		}
		p.extractors[kind] = e
	}
}

func (p *Plugins) extractorFor(kind string) (Extractor, bool) {
	e, ok := p.extractors[kind]
	return e, ok
}
