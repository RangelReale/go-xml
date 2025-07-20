package xsdruntime

import (
	"encoding/xml"
	"fmt"
)

type Resolver interface {
	Resolve(dif *DecoderInstanceFactory) (err error)
}

type FieldResolver[T, DECT any] struct {
	Value      T          `xml:"-"`
	IsResolved bool       `xml:"-"`
	XSIType    string     `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Attributes []xml.Attr `xml:",any,attr"`
	Content    string     `xml:",innerxml"`
}

func (d *FieldResolver[T, DECT]) createType(dif *DecoderInstanceFactory) (T, error) {
	instance, err := dif.CreateAliased(d.XSIType)
	if err != nil {
		var et T
		return et, err
	}
	if ctype, ok := instance.(T); ok {
		return ctype, err
	}
	var et T
	return et, fmt.Errorf("expected created type to be '%T' but is '%T'", et, instance)
}

func (d *FieldResolver[T, DECT]) Resolve(dif *DecoderInstanceFactory) error {
	if d.IsResolved {
		return nil
	}

	instance, err := d.createType(dif)
	if err != nil {
		return err
	}

	scontent := dif.Namespaces.WrapNamespacesInXML("root", d.Content, d.Attributes...)
	if err := xml.Unmarshal([]byte(scontent), &instance); err != nil {
		return err
	}

	d.Value = instance
	if valueResolver, ok := any(d.Value).(Resolver); ok {
		err = valueResolver.Resolve(dif)
		if err != nil {
			return err
		}
	}
	d.IsResolved = true

	return nil
}

func (d *FieldResolver[T, DECT]) Decode() (DECT, error) {
	var errT DECT
	if !d.IsResolved {
		return errT, fmt.Errorf("field is not resolved")
	}
	itemDec, ok := any(d.Value).(interface {
		Decode() (DECT, error)
	})
	if !ok {
		return errT, fmt.Errorf("field does not implement Decode for the expected type")
	}
	dec, err := itemDec.Decode()
	if err != nil {
		return errT, err
	}
	return dec, nil
}
