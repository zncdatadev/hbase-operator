package util

import "encoding/xml"

const (
	XMLStylesheet = `<?xml-stylesheet type="text/xsl" href="configuration.xsl"?>` + "\n"
)

type configuration struct {
	XMLName    xml.Name    `xml:"configuration"`
	Properties []*Property `xml:"property"`
}

type Property struct {
	XMLName     xml.Name `xml:"property"`
	Name        string   `xml:"name"`
	Value       string   `xml:"value"`
	Description string   `xml:"description,omitempty"`
}

type XMLConfiguration struct {
	Properties    []*Property
	XMLStylesheet string
}

func NewXMLConfiguration(
	properties []*Property,
) *XMLConfiguration {
	return &XMLConfiguration{
		Properties:    properties,
		XMLStylesheet: XMLStylesheet,
	}
}

func (x *XMLConfiguration) AddProperty(p *Property) {
	x.Properties = append(x.Properties, p)
}

func (x *XMLConfiguration) AddPropertyWithKV(name, value, description string) {
	x.AddProperty(&Property{Name: name, Value: value, Description: description})
}

func (x *XMLConfiguration) getHeader() string {
	if x.XMLStylesheet == "" {
		return xml.Header + XMLStylesheet
	}
	return xml.Header + x.XMLStylesheet
}

func (x *XMLConfiguration) Marshal() (string, error) {
	c := &configuration{Properties: x.Properties}
	data, err := xml.MarshalIndent(c, "", "    ")
	if err != nil {
		return "", err
	}

	fullXML := x.getHeader() + string(data)

	return fullXML, nil
}
