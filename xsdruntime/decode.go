package xsdruntime

import (
	"encoding/xml"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type DecoderAny interface {
	DecodeAny() (any, error)
}

type DecoderInstanceFactory struct {
	Namespaces *DecoderNamespace
	Factory    *InstanceFactory
}

func NewDecoderInstanceFactory(namespaces *DecoderNamespace, factory *InstanceFactory) *DecoderInstanceFactory {
	return &DecoderInstanceFactory{
		Namespaces: namespaces,
		Factory:    factory,
	}
}

func (f *DecoderInstanceFactory) CreateAliased(aliasedName string) (any, error) {
	alias, name, found := strings.Cut(aliasedName, ":")
	if !found {
		return f.Create(f.Namespaces.Root, aliasedName)
	}

	ns, ok := f.Namespaces.AliasNamespace(alias)
	if !ok {
		return nil, fmt.Errorf("could not find namespace for alias: %s", alias)
	}

	return f.Create(ns, name)
}

func (f *DecoderInstanceFactory) Create(ns, name string) (any, error) {
	nsf, ok := f.Factory.Get(ns)
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

func (dn *DecoderNamespace) WrapNamespacesInXML(rootElement string, content string, attrs ...xml.Attr) string {
	var data strings.Builder
	_, _ = data.WriteString(`<?xml version="1.0" encoding="utf-8"?>` + "\n" + fmt.Sprintf(`<%s`, rootElement))
	if dn.Root != "" {
		_, _ = data.WriteString(fmt.Sprintf(` xmlns="%s"`, dn.Root))
	}
	for _, nsalias := range slices.Sorted(maps.Keys(dn.Aliases)) {
		nsname := dn.Aliases[nsalias]
		_, _ = data.WriteString(fmt.Sprintf(` xmlns:%s="%s"`, nsalias, nsname))
	}
	for _, attr := range attrs {
		if attr.Name.Space == "" {
			_, _ = data.WriteString(fmt.Sprintf(` %s="%s"`, attr.Name.Local, attr.Value))
		} else {
			_, _ = data.WriteString(fmt.Sprintf(` %s:%s="%s"`, attr.Name.Space, attr.Name.Local, attr.Value))
		}
	}

	_, _ = data.WriteString(`>` + "\n")

	_, _ = data.WriteString(content)

	_, _ = data.WriteString("\n" + fmt.Sprintf(`</%s>`, rootElement))

	return data.String()
}
