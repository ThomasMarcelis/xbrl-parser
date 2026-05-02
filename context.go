package xbrl

import (
	"encoding/xml"
	"errors"
)

// Context contains information about the Entity being described, the reporting Period, and the reporting Scenario.
// All of which are necessary for understanding a business Fact captured as an XBRL item.
// Scenario is preserved as raw XML, but scenario validation and interpretation are not implemented.
// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.7
type Context struct {
	ID string `xml:"id,attr"`

	Period   Period      `xml:"period"`
	Entity   Entity      `xml:"entity"`
	Scenario *RawElement `xml:"scenario"`
}

// Entity documents the business entity for a Context (business, government department, individual, etc.).
// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.7.3
type Entity struct {
	Identifier Identifier `xml:"identifier"`
	Segments   Segments   `xml:"segment"`
}

// Validate checks that e contains the structural fields required by XBRL.
func (e Entity) Validate() error {
	if e.Identifier.Scheme == "" {
		return errors.New("entity identifier missing scheme")
	}
	if e.Identifier.Value == "" {
		return errors.New("entity identifier missing value")
	}

	return nil
}

// IsValid validates e and returns true if no error was found.
func (e Entity) IsValid() bool {
	return e.Validate() == nil
}

// Identifier specifies a scheme for identifying business entities and an identifier that follows the scheme.
// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.7.3.1
// For Example:
// <identifier scheme="http://www.sec.gov/CIK">0000320193</identifier>
//
// The above `identifier` element specifies that the scheme for identifying the entity is through SEC CIK numbers,
// and that the identifier itself is 0000320193 (the CIK for Apple Inc.).
type Identifier struct {
	Scheme string `xml:"scheme,attr"`
	Value  string `xml:",chardata"`
}

// Segments is a type alias for a slice of Segment structs.
// It implements xml.Unmarshaller and puts and unmarshals any sub-elements as a Segment and puts it into the slice.
type Segments []Segment

// Segment is an optional container for additional information used to identify a business segment more completely
// for cases where the Identifier is insufficient.
// There are no Segments defined in the base XBRL spec, they must be defined in other XML schemas.
// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.7.3.2
type Segment struct {
	XMLName    xml.Name
	Attributes []xml.Attr `xml:",any,attr"`
	Value      string     `xml:",chardata"`
	InnerXML   string     `xml:",innerxml"`
}

// UnmarshalXML implements xml.Unmarshaller for Segments.
// It unmarshals any sub-elements as Segments and puts them into this slice.
func (s *Segments) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var segmentsAnon struct {
		Segments []Segment `xml:",any"`
	}

	if err := d.DecodeElement(&segmentsAnon, &start); err != nil {
		return err
	}

	*s = segmentsAnon.Segments
	return nil
}

// PeriodType describes which supported shape a Period has.
type PeriodType string

// All the supported PeriodType values. See Period.Type() for more information.
const (
	// PeriodTypeDuration is a period with startDate and endDate.
	PeriodTypeDuration PeriodType = "duration"
	// PeriodTypeInstant is a period with instant.
	PeriodTypeInstant PeriodType = "instant"
	// PeriodTypeForever is a period with forever.
	PeriodTypeForever PeriodType = "forever"
	// PeriodTypeInvalid is a period that does not match exactly one supported shape.
	PeriodTypeInvalid PeriodType = "invalid"
)

// Period contains an instant or interval of time for a Context.
// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.7.2
type Period struct {
	// StartDate is non-nil if Period.Type() returns Duration.
	StartDate *string `xml:"startDate"`
	// EndDate is non-nil if Period.Type() returns Duration.
	EndDate *string `xml:"endDate"`

	// Instant is non-nil if Period.Type() returns Instant
	Instant *string `xml:"instant"`

	// Forever is non-nil if Period.Type() returns Forever.
	// Note ideally this would be a bool, but the XML Unmarshaller doesn't support
	// setting a boolean flag based on the existence of an empty tag (in this case `<forever/>`).
	Forever *struct{} `xml:"forever"`
}

// Type returns the type of this period to help clarify what fields in the Period struct are non-nil and valid to use.
// The comments on the attributes inside the Period struct explain when they can be used depending on what this function returns.
func (p Period) Type() PeriodType {
	periodType := PeriodTypeInvalid
	matches := 0

	if p.Forever != nil {
		periodType = PeriodTypeForever
		matches++
	}

	if p.Instant != nil {
		periodType = PeriodTypeInstant
		matches++
	}

	if p.StartDate != nil && p.EndDate != nil {
		periodType = PeriodTypeDuration
		matches++
	}

	if matches != 1 {
		return PeriodTypeInvalid
	}

	return periodType
}

// Validate checks that p has exactly one supported XBRL period shape.
func (p Period) Validate() error {
	switch p.Type() {
	case PeriodTypeDuration:
		if *p.StartDate == "" {
			return errors.New("duration period missing startDate")
		}
		if *p.EndDate == "" {
			return errors.New("duration period missing endDate")
		}
	case PeriodTypeInstant:
		if *p.Instant == "" {
			return errors.New("instant period missing value")
		}
	case PeriodTypeForever:
		return nil
	default:
		return errors.New("period must have exactly one of duration, instant, or forever")
	}

	return nil
}

// IsValid validates p and returns true if no error was found.
func (p Period) IsValid() bool {
	return p.Validate() == nil
}

// Validate checks that c contains the structural fields this parser supports.
func (c Context) Validate() error {
	if c.ID == "" {
		return errors.New("context missing id")
	}
	if err := c.Entity.Validate(); err != nil {
		return err
	}
	if err := c.Period.Validate(); err != nil {
		return err
	}
	if c.Scenario != nil {
		return errors.New("scenario is not supported")
	}

	return nil
}

// IsValid validates c and returns true if no error was found.
func (c Context) IsValid() bool {
	return c.Validate() == nil
}
