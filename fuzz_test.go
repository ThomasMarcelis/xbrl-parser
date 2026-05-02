//go:build go1.18
// +build go1.18

package xbrl

import "testing"

func FuzzParseAndValidate(f *testing.F) {
	f.Add([]byte(`<xbrl/>`))
	f.Add([]byte(`<xbrl><context id="c1"><entity><identifier scheme="s">e</identifier></entity><period><forever/></period></context></xbrl>`))
	f.Add([]byte(`<xbrl><ci:assets contextRef="missing" unitRef="u1" precision="3">727</ci:assets></xbrl>`))

	f.Fuzz(func(t *testing.T, data []byte) {
		doc, err := Parse(data)
		if err != nil {
			return
		}

		_ = doc.Validate()
	})
}
