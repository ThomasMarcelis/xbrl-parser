package xbrl_test

import (
	"encoding/xml"
	"fmt"

	"github.com/massive-com/xbrl-parser/v2"
)

const doc = `<xbrl
    xmlns="http://www.xbrl.org/2003/instance"
    xmlns:link="http://www.xbrl.org/2003/linkbase"
    xmlns:xlink="http://www.w3.org/1999/xlink"
    xmlns:ci="http://www.xbrl.org/us/gaap/ci/2003/usfr-ci-2003">
    <link:schemaRef xlink:type="simple" xlink:href="http://www.xbrl.org/us/fr/ci/2000-07-31/usfr-ci-2003.xsd"/>

    <context id="c1">
        <entity>
            <identifier scheme="http://www.sec.gov/CIK">0000320193</identifier>
        </entity>
        <period>
            <instant>2021-04-16</instant>
        </period>
    </context>

    <ci:assets precision="3" unitRef="u1" contextRef="c1">727</ci:assets>

    <unit id="u1">
        <measure>shares</measure>
    </unit>
</xbrl>`

func Example() {
	var processed xbrl.XBRL

	if err := xml.Unmarshal([]byte(doc), &processed); err != nil {
		panic(err)
	}
	if err := processed.Validate(); err != nil {
		panic(err)
	}

	fact := processed.Facts[0]
	factType := fact.Type()
	numericValue, err := fact.NumericValue()

	factContext := processed.ContextsByID[fact.ContextRef]
	factUnit := processed.UnitsByID[*fact.UnitRef]

	if err != nil {
		panic(err)
	}

	fmt.Printf("Fact: %s (namespace: %s, type: %s)\n", fact.XMLName.Local, fact.XMLName.Space, factType)
	fmt.Printf("      %.0f %s on %s\n", numericValue, factUnit.String(), *factContext.Period.Instant)

	// Output: Fact: assets (namespace: http://www.xbrl.org/us/gaap/ci/2003/usfr-ci-2003, type: non_fraction)
	//       727 shares on 2021-04-16
}

func ExampleParse() {
	processed, err := xbrl.Parse([]byte(doc))
	if err != nil {
		panic(err)
	}
	if err := processed.Validate(); err != nil {
		panic(err)
	}

	fmt.Println(len(processed.Facts))

	// Output: 1
}
