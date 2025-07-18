package xsdruntime

import "encoding/xml"

type FieldDecoder[T any] struct {
	Value      T          `xml:"-"`
	XSIType    string     `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Attributes []xml.Attr `xml:",any,attr"`
	Content    string     `xml:",innerxml"`
}
