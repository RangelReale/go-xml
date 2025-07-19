package xsdruntime

import (
	"encoding/xml"
	"fmt"
)

type FieldResolver[T any] struct {
	Value      T          `xml:"-"`
	IsResolved bool       `xml:"-"`
	XSIType    string     `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Attributes []xml.Attr `xml:",any,attr"`
	Content    string     `xml:",innerxml"`
}

func (d *FieldResolver[T]) createType(dif *DecoderInstanceFactory) (T, error) {
	instance, err := dif.CreateAliased(d.XSIType)
	if err != nil {
		var et T
		return et, err
	}
	if ctype, ok := instance.(T); ok {
		return ctype, err
	}
	var et T
	return et, fmt.Errorf("expected created type to be 'Term' but is %T", instance)
}

func (d *FieldResolver[T]) Resolve(dif *DecoderInstanceFactory) error {
	if d.IsResolved {
		return nil
	}

	instance, err := d.createType(dif)
	if err != nil {
		return err
	}

	scontent := dif.Namespaces.WrapNamespacesInXML("root", d.Content)
	if err := xml.Unmarshal([]byte(scontent), &instance); err != nil {
		return err
	}

	d.Value = instance
	d.IsResolved = true

	return nil
}
