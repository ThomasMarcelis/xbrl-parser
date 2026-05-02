package xbrl

import (
	"encoding/xml"
	"errors"
	"io"
)

// Parse unmarshals an XBRL instance document from data.
// It does not call XBRL.Validate; parsing and structural validation are separate operations.
func Parse(data []byte) (XBRL, error) {
	var doc XBRL
	if err := xml.Unmarshal(data, &doc); err != nil {
		return XBRL{}, err
	}

	return doc, nil
}

// ParseReader decodes an XBRL instance document from r using encoding/xml.
// It does not call XBRL.Validate.
func ParseReader(r io.Reader) (XBRL, error) {
	if r == nil {
		return XBRL{}, errors.New("nil reader")
	}

	return Decode(xml.NewDecoder(r))
}

// Decode decodes an XBRL instance document with decoder.
// Use this helper when callers need to configure xml.Decoder, such as setting CharsetReader.
// It does not call XBRL.Validate.
func Decode(decoder *xml.Decoder) (XBRL, error) {
	if decoder == nil {
		return XBRL{}, errors.New("nil decoder")
	}

	var doc XBRL
	if err := decoder.Decode(&doc); err != nil {
		return XBRL{}, err
	}

	return doc, nil
}
