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

type DecoderInstanceFactory struct {
	Namespaces *DecoderNamespace
	Factories  *DecoderFactoryList
}

func NewDecoderInstanceFactory(namespaces *DecoderNamespace, factories *DecoderFactoryList) *DecoderInstanceFactory {
	return &DecoderInstanceFactory{
		Namespaces: namespaces,
		Factories:  factories,
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
	nsf, ok := f.Factories.Get(ns)
	if !ok {
		return nil, fmt.Errorf("could not find factory for namespace '%s'", ns)
	}
	return nsf.NewInstance(name)
}

type DecoderFactoryList struct {
	factories map[string]InfoDecl
}

func NewDecoderFactoryList() *DecoderFactoryList {
	return &DecoderFactoryList{
		factories: make(map[string]InfoDecl),
	}
}

func (f *DecoderFactoryList) Register(infoDecl ...InfoDecl) {
	for _, i := range infoDecl {
		f.factories[i.Namespace()] = i
	}
}

func (f *DecoderFactoryList) Get(ns string) (nf InfoDecl, ok bool) {
	nf, ok = f.factories[ns]
	return
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
