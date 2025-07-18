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

type FieldDecoder[T any] struct {
	Value      T          `xml:"-"`
	XSIType    string     `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Attributes []xml.Attr `xml:",any,attr"`
	Content    string     `xml:",innerxml"`
}

type DecoderInstanceFactoryFunc func(name string) (any, error)

type DecoderInstanceFactory struct {
	Namespaces *DecoderNamespace
	Factories  map[string]DecoderInstanceFactoryFunc
}

func NewDecoderInstanceFactory(namespaces *DecoderNamespace) *DecoderInstanceFactory {
	return &DecoderInstanceFactory{
		Namespaces: namespaces,
		Factories:  make(map[string]DecoderInstanceFactoryFunc),
	}
}

func (f *DecoderInstanceFactory) Register(namespace string, factory func(name string) (any, error)) {
	f.Factories[namespace] = factory
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
	return nsf(name)
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

	_, _ = data.WriteString("\n" + `</otx>`)

	return data.String()
}
