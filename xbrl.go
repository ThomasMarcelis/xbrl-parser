package xbrl

import (
	"encoding/xml"
	"fmt"
)

const (
	xbrlInstanceNamespace = "http://www.xbrl.org/2003/instance"
)

// NotImplemented represents a count of expected XBRL elements that are not handled in detail.
// New code should prefer the RawElement fields that preserve XML names, attributes, and inner XML.
type NotImplemented []*struct{}

// RawElement preserves an XML element that this package does not model in detail.
type RawElement struct {
	XMLName    xml.Name
	Attributes []xml.Attr `xml:",any,attr"`
	InnerXML   string     `xml:",innerxml"`
}

// RawXBRL represents the XML structure of an XBRL document.
// This is not a feature complete XBRL parser!
// See the fields of type RawElement and NotImplemented for an idea of what's missing.
// Also note that this struct doesn't support Tuple facts (https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.9)
//
// You can use this struct directly, but XBRL is structured in a more convenient way.
// See the comment on XBRL for more info.
type RawXBRL struct {
	XMLName    xml.Name
	Attributes []xml.Attr

	Contexts []Context `xml:"context"`
	Units    []Unit    `xml:"unit"`

	Facts []Fact `xml:",any"`

	SchemaRefs          []RawElement
	LinkbaseRefs        []RawElement
	RoleRefs            []RawElement
	ArcRoleRefs         []RawElement
	FootnoteLinks       []RawElement
	UnsupportedTopLevel []RawElement

	// The fields below are not properly implemented, but need to be here so they aren't lumped into the `Facts` slice.

	SchemaRef    NotImplemented `xml:"schemaRef"`
	LinkbaseRef  NotImplemented `xml:"linkbaseRef"`
	RoleRef      NotImplemented `xml:"roleRef"`
	ArcRoleRef   NotImplemented `xml:"arcroleRef"`
	FootnoteLink NotImplemented `xml:"footnoteLink"`
}

// UnmarshalXML implements xml.Unmarshaler and preserves the XBRL root envelope
// while still routing taxonomy-defined top-level elements into Facts.
func (r *RawXBRL) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*r = RawXBRL{
		XMLName:    start.Name,
		Attributes: copyAttrs(start.Attr),
	}

	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch token := token.(type) {
		case xml.StartElement:
			if err := r.decodeChild(d, token); err != nil {
				return err
			}
		case xml.EndElement:
			if token.Name == start.Name {
				return nil
			}
		}
	}
}

func (r *RawXBRL) decodeChild(d *xml.Decoder, start xml.StartElement) error {
	switch start.Name.Local {
	case "context":
		var context Context
		if err := d.DecodeElement(&context, &start); err != nil {
			return err
		}
		r.Contexts = append(r.Contexts, context)
	case "unit":
		var unit Unit
		if err := d.DecodeElement(&unit, &start); err != nil {
			return err
		}
		r.Units = append(r.Units, unit)
	case "schemaRef":
		element, err := decodeRawElement(d, start)
		if err != nil {
			return err
		}
		r.SchemaRefs = append(r.SchemaRefs, element)
		r.SchemaRef = append(r.SchemaRef, &struct{}{})
	case "linkbaseRef":
		element, err := decodeRawElement(d, start)
		if err != nil {
			return err
		}
		r.LinkbaseRefs = append(r.LinkbaseRefs, element)
		r.LinkbaseRef = append(r.LinkbaseRef, &struct{}{})
	case "roleRef":
		element, err := decodeRawElement(d, start)
		if err != nil {
			return err
		}
		r.RoleRefs = append(r.RoleRefs, element)
		r.RoleRef = append(r.RoleRef, &struct{}{})
	case "arcroleRef":
		element, err := decodeRawElement(d, start)
		if err != nil {
			return err
		}
		r.ArcRoleRefs = append(r.ArcRoleRefs, element)
		r.ArcRoleRef = append(r.ArcRoleRef, &struct{}{})
	case "footnoteLink":
		element, err := decodeRawElement(d, start)
		if err != nil {
			return err
		}
		r.FootnoteLinks = append(r.FootnoteLinks, element)
		r.FootnoteLink = append(r.FootnoteLink, &struct{}{})
	default:
		if isKnownUnsupportedTopLevel(start.Name) {
			element, err := decodeRawElement(d, start)
			if err != nil {
				return err
			}
			r.UnsupportedTopLevel = append(r.UnsupportedTopLevel, element)
			return nil
		}

		var fact Fact
		if err := d.DecodeElement(&fact, &start); err != nil {
			return err
		}
		r.Facts = append(r.Facts, fact)
	}

	return nil
}

func isKnownUnsupportedTopLevel(name xml.Name) bool {
	return name.Space == xbrlInstanceNamespace && (name.Local == "item" || name.Local == "tuple")
}

func decodeRawElement(d *xml.Decoder, start xml.StartElement) (RawElement, error) {
	var element RawElement
	if err := d.DecodeElement(&element, &start); err != nil {
		return RawElement{}, err
	}

	return element, nil
}

func copyAttrs(attrs []xml.Attr) []xml.Attr {
	if len(attrs) == 0 {
		return nil
	}

	copied := make([]xml.Attr, len(attrs))
	copy(copied, attrs)
	return copied
}

func copyRawElements(elements []RawElement) []RawElement {
	if len(elements) == 0 {
		return nil
	}

	copied := make([]RawElement, len(elements))
	for index, element := range elements {
		copied[index] = element
		copied[index].Attributes = copyAttrs(element.Attributes)
	}

	return copied
}

// XBRL contains raw context and unit slices plus maps so contexts and units can be accessed easier when looping through facts.
// You can either unmarshal XML directly into this struct (it has a custom unmarshaller),
// or you can unmarshal XML into a RawXBRL struct and call NewProcessedXBRL(RawXBRL) to process the raw XBRL into this format.
type XBRL struct {
	XMLName    xml.Name
	Attributes []xml.Attr

	Contexts []Context
	Units    []Unit

	ContextsByID map[string]Context
	UnitsByID    map[string]Unit

	Facts []Fact

	SchemaRefs          []RawElement
	LinkbaseRefs        []RawElement
	RoleRefs            []RawElement
	ArcRoleRefs         []RawElement
	FootnoteLinks       []RawElement
	UnsupportedTopLevel []RawElement
}

// NewProcessedXBRL constructs a XBRL struct from a RawXBRL struct.
func NewProcessedXBRL(raw RawXBRL) XBRL {
	contextsByID := make(map[string]Context, len(raw.Contexts))
	unitsByID := make(map[string]Unit, len(raw.Units))

	for _, context := range raw.Contexts {
		contextsByID[context.ID] = context
	}

	for _, unit := range raw.Units {
		unitsByID[unit.ID] = unit
	}

	return XBRL{
		XMLName:             raw.XMLName,
		Attributes:          copyAttrs(raw.Attributes),
		Contexts:            append([]Context(nil), raw.Contexts...),
		Units:               append([]Unit(nil), raw.Units...),
		ContextsByID:        contextsByID,
		UnitsByID:           unitsByID,
		Facts:               append([]Fact(nil), raw.Facts...),
		SchemaRefs:          copyRawElements(raw.SchemaRefs),
		LinkbaseRefs:        copyRawElements(raw.LinkbaseRefs),
		RoleRefs:            copyRawElements(raw.RoleRefs),
		ArcRoleRefs:         copyRawElements(raw.ArcRoleRefs),
		FootnoteLinks:       copyRawElements(raw.FootnoteLinks),
		UnsupportedTopLevel: copyRawElements(raw.UnsupportedTopLevel),
	}
}

// UnmarshalXML implements xml.Unmarshaler and unmarshals the contents as a RawXBRL,
// then processes it and populates this struct's fields.
func (x *XBRL) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw RawXBRL
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}

	*x = NewProcessedXBRL(raw)
	return nil
}

// Validate checks basic XBRL structure: contexts, units, facts, duplicate IDs, and fact references.
// It does not perform taxonomy-aware, accounting, linkbase, footnote, tuple, or scenario validation.
func (x XBRL) Validate() error {
	contextsByID, err := x.validatedContextsByID()
	if err != nil {
		return err
	}

	unitsByID, err := x.validatedUnitsByID()
	if err != nil {
		return err
	}

	if len(x.UnsupportedTopLevel) > 0 {
		element := x.UnsupportedTopLevel[0]
		return fmt.Errorf("unsupported top-level element: %s:%s", element.XMLName.Space, element.XMLName.Local)
	}

	for _, fact := range x.Facts {
		if err := fact.Validate(); err != nil {
			return fmt.Errorf("invalid fact (%s:%s): %w", fact.XMLName.Space, fact.XMLName.Local, err)
		}

		if _, exists := contextsByID[fact.ContextRef]; !exists {
			return fmt.Errorf("fact (%s:%s) references non-existent context: %s", fact.XMLName.Space, fact.XMLName.Local, fact.ContextRef)
		}

		if fact.UnitRef != nil {
			if _, exists := unitsByID[*fact.UnitRef]; !exists {
				return fmt.Errorf("fact (%s:%s) references non-existent unit: %s", fact.XMLName.Space, fact.XMLName.Local, *fact.UnitRef)
			}
		}
	}

	return nil
}

// IsValid validates this struct and returns true if no error was found.
func (x XBRL) IsValid() bool {
	return x.Validate() == nil
}

func (x XBRL) validatedContextsByID() (map[string]Context, error) {
	if len(x.Contexts) == 0 {
		return x.ContextsByID, nil
	}

	contextsByID := make(map[string]Context, len(x.Contexts))
	for _, context := range x.Contexts {
		if err := context.Validate(); err != nil {
			return nil, fmt.Errorf("invalid context (%s): %w", context.ID, err)
		}
		if _, exists := contextsByID[context.ID]; exists {
			return nil, fmt.Errorf("duplicate context id: %s", context.ID)
		}

		contextsByID[context.ID] = context
	}

	return contextsByID, nil
}

func (x XBRL) validatedUnitsByID() (map[string]Unit, error) {
	if len(x.Units) == 0 {
		return x.UnitsByID, nil
	}

	unitsByID := make(map[string]Unit, len(x.Units))
	for _, unit := range x.Units {
		if err := unit.Validate(); err != nil {
			return nil, fmt.Errorf("invalid unit (%s): %w", unit.ID, err)
		}
		if _, exists := unitsByID[unit.ID]; exists {
			return nil, fmt.Errorf("duplicate unit id: %s", unit.ID)
		}

		unitsByID[unit.ID] = unit
	}

	return unitsByID, nil
}
