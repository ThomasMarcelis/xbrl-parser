package xbrl

import (
	"encoding/xml"
	"errors"
	"strconv"
)

// FactType describes the structural category of a parsed Fact.
type FactType string

const (
	// FactTypeNil is a fact which has an `xsi:nil` attribute set to a truthy value.
	// A nil fact is only guaranteed to have an XMLName and ContextRef.
	FactTypeNil FactType = "nil"

	// FactTypeNonNumeric is a non-nil fact that does not describe a numeric value (ie text, dates, encoded binary data, etc).
	// A non-numeric fact is guaranteed to have an XMLName, ContextRef, and ValueStr.
	FactTypeNonNumeric FactType = "non_numeric"

	// FactTypeNonFraction is a non-nil fact describing a numeric value that can be precisely expressed as a simple value.
	// A non-fraction fact is guaranteed to have an XMLName, ContextRef, UnitRef, ValueStr, and exactly one of Precision or Decimals.
	//
	// For example: <ci:capitalLeases contextRef="c1" unitRef="u1" precision="3">727432</ci:capitalLeases>
	//
	// Use Fact.NumericValue() for easy access to the numeric value as a float64.
	FactTypeNonFraction FactType = "non_fraction"

	// FactTypeFraction is a non-nil fact describing a numeric value that is the result of a numerator / denominator.
	// Usually the numeric value that these facts describe cannot be precisely expressed by a float64 (ie 1/3 = 0.3333...)
	// A fraction fact is guaranteed to have an XMLName, ContextRef, UnitRef, Numerator, and Denominator.
	//
	// For example:
	// <myTaxonomy:oneThird id="oneThird" unitRef="u1" contextRef="numC1">
	//     <numerator>1</numerator>
	//     <denominator>3</denominator>
	// </myTaxonomy:oneThird>
	//
	// Use Fact.NumericValue() for easy access to the numeric value as a float64,
	// but be aware that the float64 representation may not be able to precisely represent the Facts actual value.
	FactTypeFraction FactType = "fraction"
)

// ErrNonNumericFactType is returned when a fact is expected to be numeric, but is not.
var ErrNonNumericFactType = errors.New("fact is not of type FactTypeFraction or FactTypeNonFraction")

// Fact represents an item in an XBRL document.
// A Fact is a single value which is tied to a context that gives the fact more meaning.
//
// This struct contains fields that may or may not be nil depending on what type of Fact you're dealing with.
// See Fact.Type() to determine what type of fact you're dealing with, and Fact.IsValid() to be confident all the expected fields exist.
// Then the various FactTypes to understand what fields in this struct are expected to exist for each FactType.
//
// For general information and details on XBRL Facts, see here:
// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.6
type Fact struct {
	XMLName xml.Name

	// ID uniquely identifies a fact within an XBRL document.
	// The spec does not require an ID attribute, but many items have an ID attribute, which is why it's included in this model.
	ID string `xml:"id,attr"`

	// Nil is an attribute denoting whether or not this fact is expressed as nil.
	Nil *bool `xml:"nil,attr"`

	// ContextRef is the ID of the context in the XBRL document that gives more meaning to this fact.
	ContextRef string `xml:"contextRef,attr"`

	// UnitRef is the ID of the unit in the XBRL document that this fact is expressed in.
	// It is non-nil for numeric facts only.
	UnitRef *string `xml:"unitRef,attr"`

	// Precision conveys the arithmetic precision of a measurement.
	// It can be either a non-negative integer or the special value "INF", which represents infinite precision.
	// If this is a numeric fact but NOT a fraction type, Precision will be non-nil if Decimals is nil,
	//
	// Examples and more info here:
	// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.6.4
	Precision *string `xml:"precision,attr"`

	// Decimals specifies the number of decimal places to which the value of the fact represented may be considered accurate.
	// It can be either an integer (positive or negative) or the special value "INF", which represents accuracy to infinite decimal places.
	// If this is a numeric fact but NOT a fraction type, Decimals will be non-nil if Precision is nil,
	//
	// Examples and more info here:
	// https://www.xbrl.org/Specification/XBRL-2.1/REC-2003-12-31/XBRL-2.1-REC-2003-12-31+corrected-errata-2013-02-20.html#_4.6.5
	Decimals *string `xml:"decimals,attr"`

	// ValueStr will be non-nil unless this is a numeric fraction type Fact.
	// Use NumericValue() to easily get the numeric value that this fact represents, regardless of whether or not it's a fraction type.
	ValueStr *string `xml:",chardata"`

	// Numerator and Denominator will be non-nil values if this is a fraction type Fact.
	// Use NumericValue() to easily get the numeric value that this fact represents, regardless of whether or not it's a fraction type.
	Numerator   *float64 `xml:"numerator"`
	Denominator *float64 `xml:"denominator"`
}

// Type returns the type of this Fact. See the comments on the various FactTypes for more information.
// Note that this function returning a particular type does not necessarily mean that the fact is semantically correct.
// See IsValid() to be certain that the fact is valid.
func (f Fact) Type() FactType {
	// If the nil attribute exists and is true, this is simply a nil fact
	if f.Nil != nil && *f.Nil {
		return FactTypeNil
	}

	// If the unitRef attribute exists, this is some kind of numeric attribute
	if f.UnitRef != nil {
		// If we have a numerator and denominator, it's a fraction fact
		if f.Numerator != nil && f.Denominator != nil {
			return FactTypeFraction
		}

		// Otherwise it's a simple non fraction numeric type.
		return FactTypeNonFraction
	}

	// All that's left is a plain non-numeric fact type
	return FactTypeNonNumeric
}

// IsValid confirms that f has the structural fields required for its FactType.
func (f Fact) IsValid() bool {
	return f.Validate() == nil
}

// Validate checks that f has the structural fields required for its FactType.
// It does not perform taxonomy-aware validation.
func (f Fact) Validate() error {
	// All facts must have a context ref
	if f.ContextRef == "" {
		return errors.New("missing contextRef")
	}

	// Some types have particular rules beyond what Type() checks for that must be true to be considered valid.
	switch f.Type() {
	case FactTypeFraction:
		if f.UnitRef == nil || *f.UnitRef == "" {
			return errors.New("fraction fact missing unitRef")
		}
		if f.Numerator == nil {
			return errors.New("fraction fact missing numerator")
		}
		if f.Denominator == nil {
			return errors.New("fraction fact missing denominator")
		}
		if *f.Denominator == 0 {
			return errors.New("fraction fact denominator is zero")
		}
		if f.Precision != nil || f.Decimals != nil {
			return errors.New("fraction fact cannot have precision or decimals")
		}
	case FactTypeNonFraction:
		if f.UnitRef == nil || *f.UnitRef == "" {
			return errors.New("non-fraction fact missing unitRef")
		}
		if f.ValueStr == nil {
			return errors.New("non-fraction fact missing value")
		}
		if f.Numerator != nil || f.Denominator != nil {
			return errors.New("non-fraction fact cannot have numerator or denominator")
		}
		if (f.Precision == nil) == (f.Decimals == nil) {
			return errors.New("non-fraction fact must have exactly one of precision or decimals")
		}
		if f.Precision != nil && !isValidPrecision(*f.Precision) {
			return errors.New("non-fraction fact has invalid precision")
		}
		if f.Decimals != nil && !isValidDecimals(*f.Decimals) {
			return errors.New("non-fraction fact has invalid decimals")
		}
		if _, err := strconv.ParseFloat(*f.ValueStr, 64); err != nil {
			return err
		}
	case FactTypeNonNumeric:
		if f.ValueStr == nil {
			return errors.New("non-numeric fact missing value")
		}
		if f.UnitRef != nil {
			return errors.New("non-numeric fact cannot have unitRef")
		}
		if f.Precision != nil || f.Decimals != nil {
			return errors.New("non-numeric fact cannot have precision or decimals")
		}
		if f.Numerator != nil || f.Denominator != nil {
			return errors.New("non-numeric fact cannot have numerator or denominator")
		}
	}

	return nil
}

func isValidPrecision(precision string) bool {
	if precision == "INF" {
		return true
	}

	value, err := strconv.Atoi(precision)
	return err == nil && value >= 0
}

func isValidDecimals(decimals string) bool {
	if decimals == "INF" {
		return true
	}

	_, err := strconv.Atoi(decimals)
	return err == nil
}

// NumericValue attempts to return the numeric value this fact represents.
// If this fact is a fraction type, this function returns the value of numerator / denominator.
// Note that fraction type facts generally cannot be precisely represented as a float64 and may have some rounding error.
func (f Fact) NumericValue() (float64, error) {
	switch f.Type() {
	case FactTypeFraction:
		if f.Numerator == nil {
			return 0, errors.New("fraction fact missing numerator")
		}
		if f.Denominator == nil {
			return 0, errors.New("fraction fact missing denominator")
		}
		if *f.Denominator == 0 {
			return 0, errors.New("fraction fact denominator is zero")
		}
		return *f.Numerator / *f.Denominator, nil
	case FactTypeNonFraction:
		if f.ValueStr == nil {
			return 0, errors.New("non-fraction fact missing value")
		}
		return strconv.ParseFloat(*f.ValueStr, 64)
	default:
		return 0, ErrNonNumericFactType
	}
}

// Value returns the ValueStr of this Fact, or empty string if f.ValueStr is nil.
func (f Fact) Value() string {
	if f.ValueStr != nil {
		return *f.ValueStr
	}

	return ""
}
