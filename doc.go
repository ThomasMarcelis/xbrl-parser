// Package xbrl parses XBRL 2.1 instance documents into simple Go data.
//
// The package preserves XBRL concepts such as facts, contexts, periods,
// entities, segments, units, XML names, attributes, and raw reference elements.
// It does not load taxonomies, resolve linkbases, normalize financial
// statements, transform Inline XBRL, or perform accounting-rule validation.
//
// XML unmarshalling is a first-class API:
//
//	var doc xbrl.XBRL
//	err := xml.Unmarshal(data, &doc)
//
// Parse, ParseReader, and Decode are convenience helpers around the same
// encoding/xml path. Parsing and validation are separate operations; call
// XBRL.Validate when you need structural checks for contexts, units, facts, and
// references.
package xbrl
