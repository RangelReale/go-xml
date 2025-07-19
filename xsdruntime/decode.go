package xsdruntime

import (
	"encoding/xml"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// type RootDecoder[T any] struct {
// 	Attrs []xml.Attr `xml:",any,attr"`
// 	T // not allowed by Go
// }

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

type DecoderInstanceFactory struct {
	Namespaces *DecoderNamespace
	Factories  map[string]InfoDecl
}

func NewDecoderInstanceFactory(namespaces *DecoderNamespace) *DecoderInstanceFactory {
	return &DecoderInstanceFactory{
		Namespaces: namespaces,
		Factories:  make(map[string]InfoDecl),
	}
}

func (f *DecoderInstanceFactory) Register(infoDecl ...InfoDecl) {
	for _, i := range infoDecl {
		f.Factories[i.Namespace()] = i
	}
}

func (f *DecoderInstanceFactory) CreateAliased(aliasedName string) (any, error) {
	alias, name, found := strings.Cut(aliasedName, ":")
	if !found {
		return nil, fmt.Errorf("name does not have an alias: %s", aliasedName)
	}

	ns, ok := f.Namespaces.AliasNamespace(alias)
	if !ok {
		return nil, fmt.Errorf("could not find namespace for alias: %s", alias)
	}

	return f.Create(ns, name)
}

func (f *DecoderInstanceFactory) Create(ns, name string) (any, error) {
	nsf, ok := f.Factories[ns]
	if !ok {
		return nil, fmt.Errorf("could not find factory for namespace '%s'", ns)
	}
	return nsf.NewInstance(name)
}

type DecoderNamespace struct {
	Root    string
	Aliases map[string]string
}

func NewDecoderNamespace(attrs []xml.Attr) *DecoderNamespace {
	ret := &DecoderNamespace{
		Aliases: make(map[string]string),
	}
	for _, attr := range attrs {
		if attr.Name.Local == "xmlns" && attr.Name.Space == "" {
			ret.Root = attr.Value
			continue
		}
		if attr.Name.Space != "xmlns" {
			continue
		}
		ret.Aliases[attr.Name.Local] = attr.Value
	}
	return ret
}

func (dn *DecoderNamespace) AliasNamespace(alias string) (v string, ok bool) {
	v, ok = dn.Aliases[alias]
	return
}

func (dn *DecoderNamespace) WrapNamespacesInXML(rootElement string, content string) string {
	var data strings.Builder
	_, _ = data.WriteString(`<?xml version="1.0" encoding="utf-8"?>` + "\n" + fmt.Sprintf(`<%s`, rootElement))
	if dn.Root != "" {
		_, _ = data.WriteString(fmt.Sprintf(` xmlns="%s"`, dn.Root))
	}
	for _, nsalias := range slices.Sorted(maps.Keys(dn.Aliases)) {
		nsname := dn.Aliases[nsalias]
		_, _ = data.WriteString(fmt.Sprintf(` xmlns:%s="%s"`, nsalias, nsname))
	}

	_, _ = data.WriteString(`>` + "\n")

	_, _ = data.WriteString(content)

	_, _ = data.WriteString("\n" + fmt.Sprintf(`</%s>`, rootElement))

	return data.String()
}
