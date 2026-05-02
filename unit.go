package xbrl

import (
	"errors"
	"strings"
)

// Unit specifies the unit in which a numeric fact has been measured.
// A Unit can be either a simple measure, product of measures, or a ratio of products of measures with a numerator and a denominator.
//
// A simple unit that represents shares looks like:
// <unit>
//
//	<measure>shares</measure>
//
// </unit>
//
// Numeric Facts reference units by ID via the Fact's `unitRef` attribute.
// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.8
type Unit struct {
	ID       string   `xml:"id,attr"`
	Measures Measures `xml:"measure"`
	Divide   *Divide  `xml:"divide"`
}

// Divide represents a ratio of units that has a numerator and a denominator.
// For example, XBRL can represent a complex unit like earnings per share (EPS) as dollars per share (USD / share):
// <unit>
//
//	    <divide>
//		       <unitNumerator>
//	            <measure>iso4217:USD</measure>
//	        </unitNumerator>
//	        <unitDenominator>
//	            <measure>shares</measure>
//	        </unitDenominator>
//	    </divide>
//
// </unit>
//
// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.8.2
type Divide struct {
	Numerator   Measures `xml:"unitNumerator>measure"`
	Denominator Measures `xml:"unitDenominator>measure"`
}

// Measure represents a unit of measure. The element value can be xml namespaced (xsd:Qname) or as plain text.
// XML namespaced: <measure>iso4217:USD</measure>
// plain text:     <measure>shares</measure>
//
// Note that if the value is XML namespaced, the namespace should be declared in the XML, but this parser does not validate that.
// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.8.2
type Measure struct {
	Value string `xml:",chardata"`
}

// Measures is a product of one or more Measure values.
type Measures []Measure

// Validate checks that u has the structural fields required by XBRL.
func (u Unit) Validate() error {
	if u.ID == "" {
		return errors.New("unit missing id")
	}
	if (len(u.Measures) == 0) == (u.Divide == nil) {
		return errors.New("unit must have either measures or divide")
	}

	if u.Divide != nil {
		return u.Divide.Validate()
	}

	return u.Measures.Validate()
}

// IsValid validates u and returns true if no error was found.
func (u Unit) IsValid() bool {
	return u.Validate() == nil
}

// Validate checks that d has numerator and denominator measures.
func (d Divide) Validate() error {
	if len(d.Numerator) == 0 {
		return errors.New("divide missing numerator measures")
	}
	if len(d.Denominator) == 0 {
		return errors.New("divide missing denominator measures")
	}
	if err := d.Numerator.Validate(); err != nil {
		return err
	}
	if err := d.Denominator.Validate(); err != nil {
		return err
	}

	return nil
}

// IsValid validates d and returns true if no error was found.
func (d Divide) IsValid() bool {
	return d.Validate() == nil
}

// Validate checks that m contains a non-empty measure value.
func (m Measure) Validate() error {
	if m.Value == "" {
		return errors.New("measure missing value")
	}

	return nil
}

// IsValid validates m and returns true if no error was found.
func (m Measure) IsValid() bool {
	return m.Validate() == nil
}

// Validate checks that m has at least one measure and that each measure has a value.
func (m Measures) Validate() error {
	if len(m) == 0 {
		return errors.New("measures missing values")
	}

	for _, measure := range m {
		if err := measure.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// IsValid validates m and returns true if no error was found.
func (m Measures) IsValid() bool {
	return m.Validate() == nil
}

// String returns a human readable representation of the Unit.
func (u Unit) String() string {
	// If the Divide element is not nil, there can be no top-level Measures.
	if u.Divide != nil {
		return u.Divide.Numerator.String() + " / " + u.Divide.Denominator.String()
	}

	// If the Divide element is nil, there must be 1+ top-level Measures.
	return u.Measures.String()
}

// String returns the local name of the measure if the value is formatted as 'xsd:Qname', otherwise the value itself is returned.
// This is a display helper only. Use Measure.Value when the raw XBRL value is significant.
// Ex: `<measure>iso4217:USD</measure>` -> "USD"
//
//	`<measure>shares</measure>`      -> "shares"
func (m Measure) String() string {
	if index := strings.IndexRune(m.Value, ':'); index != -1 && index < len(m.Value) {
		return m.Value[index+1 : len(m.Value)]
	}

	return m.Value
}

// String returns a human readable representation of the product of all the `Measure`s in this slice.
func (m Measures) String() string {
	// More than one Measure implies multiplication.
	var builder strings.Builder
	for index, measure := range m {
		if index > 0 {
			builder.WriteString(" * ")
		}

		builder.WriteString(measure.String())
	}

	return builder.String()
}
