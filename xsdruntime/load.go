package xsdruntime

import (
	"encoding/xml"
	"io"
	"slices"
)

func Load[T Resolver](r io.Reader, factories *InstanceFactory) (T, error) {
	l := &loadType[T]{}
	if err := xml.NewDecoder(r).Decode(l); err != nil {
		var el T
		return el, err
	}
	dif := NewDecoderInstanceFactory(NewDecoderNamespace(l.Attrs), factories)
	if err := l.Value.Resolve(dif); err != nil {
		var el T
		return el, err
	}
	return l.Value, nil
}

type loadType[T any] struct {
	Attrs []xml.Attr `xml:",any,attr"`
	Value T
}

func (l *loadType[T]) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	l.Attrs = slices.Clone(start.Attr)
	return d.DecodeElement(&l.Value, &start)
}
