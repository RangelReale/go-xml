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
