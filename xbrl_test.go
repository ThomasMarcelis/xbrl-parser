package xbrl

import (
	"encoding/xml"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalXBRL(t *testing.T) {
	t.Run("real-world xbrl from 2021", func(t *testing.T) {
		f, err := os.Open("test_data/aapl-20210327_htm.xml")
		require.NoError(t, err)
		defer f.Close()

		var content XBRL
		decoder := xml.NewDecoder(f)

		require.NoError(t, decoder.Decode(&content))
		require.NoError(t, content.Validate())

		// Not using assert.Len here because with the output is painful for large maps/slices
		assert.Equal(t, 283, len(content.ContextsByID))
		assert.Equal(t, 9, len(content.UnitsByID))
		assert.Equal(t, 1070, len(content.Facts))
		require.Len(t, content.SchemaRefs, 1)
		assert.Equal(t, "aapl-20210327.xsd", requireAttr(t, content.SchemaRefs[0].Attributes, "http://www.w3.org/1999/xlink", "href"))

		durationContext := content.ContextsByID["i02c0f3e92d75432fbe3c6a24022bf7b0_D20200927-20210327"]
		assert.Equal(t, PeriodTypeDuration, durationContext.Period.Type())
		assert.Equal(t, "2020-09-27", *durationContext.Period.StartDate)
		assert.Equal(t, "2021-03-27", *durationContext.Period.EndDate)

		segmentedContext := content.ContextsByID["iff44040cd61344d085f7a2b7a1076cb1_D20200927-20210327"]
		require.Len(t, segmentedContext.Entity.Segments, 1)
		assert.Equal(t, xml.Name{Space: "http://xbrl.org/2006/xbrldi", Local: "explicitMember"}, segmentedContext.Entity.Segments[0].XMLName)
		assert.Equal(t, "us-gaap:CommonStockMember", segmentedContext.Entity.Segments[0].Value)
		assert.Equal(t, "us-gaap:StatementClassOfStockAxis", requireAttr(t, segmentedContext.Entity.Segments[0].Attributes, "", "dimension"))

		usdPerShare := content.UnitsByID["usdPerShare"]
		require.NotNil(t, usdPerShare.Divide)
		assert.Equal(t, "iso4217:USD", usdPerShare.Divide.Numerator[0].Value)
		assert.Equal(t, "USD / shares", usdPerShare.String())

		eps := requireFact(t, content.Facts, "http://fasb.org/us-gaap/2020-01-31", "EarningsPerShareBasic", "1.41")
		assert.Equal(t, "ia09408265617434fbc06a7e4c6b101bc_D20201227-20210327", eps.ContextRef)
		require.NotNil(t, eps.UnitRef)
		assert.Equal(t, "usdPerShare", *eps.UnitRef)
	})

	t.Run("real-world xbrl from 2004", func(t *testing.T) {
		f, err := os.Open("test_data/edgr-2004_10k.xml") // The very first XBRL submission to the SEC!
		require.NoError(t, err)
		defer f.Close()

		var content XBRL
		decoder := xml.NewDecoder(f)
		decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
			return input, nil
		}

		require.NoError(t, decoder.Decode(&content))
		require.NoError(t, content.Validate())

		// Not using assert.Len here because with the output is painful for large maps/slices
		assert.Equal(t, 4, len(content.ContextsByID))
		assert.Equal(t, 2, len(content.UnitsByID))
		assert.Equal(t, 154, len(content.Facts))
		require.Len(t, content.SchemaRefs, 1)
		assert.Equal(t, "edgr-20050228.xsd", requireAttr(t, content.SchemaRefs[0].Attributes, "http://www.w3.org/1999/xlink", "href"))

		firstFact := content.Facts[0]
		assert.Equal(t, xml.Name{Space: "http://www.xbrl.org/us/fr/common/pte/2005-02-28", Local: "AccountsPayable"}, firstFact.XMLName)
		assert.Equal(t, "edgr_4473_inst_YTD_20041231", firstFact.ContextRef)
		require.NotNil(t, firstFact.UnitRef)
		assert.Equal(t, "USD", *firstFact.UnitRef)
		assert.Equal(t, "995000", firstFact.Value())

		instantContext := content.ContextsByID["edgr_4473_inst_YTD_20041231"]
		assert.Equal(t, PeriodTypeInstant, instantContext.Period.Type())
		assert.Equal(t, "2004-12-31", *instantContext.Period.Instant)
		assert.Equal(t, "USD", content.UnitsByID["USD"].String())
	})

	t.Run("simple xbrl happy path", func(t *testing.T) {
		xbrlBytes, err := os.ReadFile("test_data/simple_xbrl.xml")
		require.NoError(t, err)

		var content XBRL

		require.NoError(t, xml.Unmarshal(xbrlBytes, &content))
		require.NoError(t, content.Validate())

		assert.Equal(t, xml.Name{Space: "http://www.xbrl.org/2003/instance", Local: "xbrl"}, content.XMLName)
		assert.NotEmpty(t, content.Attributes)
		require.Len(t, content.SchemaRefs, 1)
		assert.Equal(t, xml.Name{Space: "http://www.xbrl.org/2003/linkbase", Local: "schemaRef"}, content.SchemaRefs[0].XMLName)
		assert.Len(t, content.Contexts, 1)
		assert.Len(t, content.Units, 1)

		require.Len(t, content.ContextsByID, 1)
		expectedContext := Context{
			ID: "c1",
			Period: Period{
				Instant: stringPtr("2021-04-16"),
			},
			Entity: Entity{
				Identifier: Identifier{
					Scheme: "http://www.sec.gov/CIK",
					Value:  "0000320193",
				},
			},
		}

		assert.Equal(t, expectedContext, content.ContextsByID["c1"])

		require.Len(t, content.UnitsByID, 1)
		expectedUnit := Unit{
			ID:       "u1",
			Measures: Measures{{Value: "shares"}},
		}

		assert.Equal(t, expectedUnit, content.UnitsByID["u1"])

		require.Len(t, content.Facts, 2)
		expectedFacts := []Fact{
			{
				XMLName:    xml.Name{Space: "http://www.xbrl.org/us/gaap/ci/2003/usfr-ci-2003", Local: "assets"},
				ContextRef: "c1",
				UnitRef:    stringPtr("u1"),
				Precision:  stringPtr("3"),
				ValueStr:   stringPtr("727"),
			},
			{
				XMLName:    xml.Name{Space: "fakens", Local: "textItem"},
				ContextRef: "c1",
				ValueStr:   stringPtr("this is a text item"),
			},
		}

		assert.Equal(t, expectedFacts, content.Facts)
	})

	t.Run("invalid xbrl", func(t *testing.T) {
		xbrlBytes, err := os.ReadFile("test_data/invalid_xbrl.xml")
		require.NoError(t, err)

		var content XBRL

		require.NoError(t, xml.Unmarshal(xbrlBytes, &content))
		assert.Error(t, content.Validate())
	})
}

func TestUnmarshalRawXBRLPreservesEnvelope(t *testing.T) {
	xbrlBytes, err := os.ReadFile("test_data/simple_xbrl.xml")
	require.NoError(t, err)

	var raw RawXBRL
	require.NoError(t, xml.Unmarshal(xbrlBytes, &raw))

	assert.Equal(t, xml.Name{Space: "http://www.xbrl.org/2003/instance", Local: "xbrl"}, raw.XMLName)
	assert.NotEmpty(t, raw.Attributes)
	require.Len(t, raw.SchemaRefs, 1)
	assert.Equal(t, xml.Name{Space: "http://www.xbrl.org/2003/linkbase", Local: "schemaRef"}, raw.SchemaRefs[0].XMLName)
	assert.NotEmpty(t, raw.SchemaRefs[0].Attributes)
	assert.Len(t, raw.SchemaRef, 1)
	assert.Len(t, raw.Facts, 2)
}

func TestUnmarshalRawXBRLPreservesReferenceAndUnsupportedElements(t *testing.T) {
	// language=xml
	doc := `<xbrl xmlns="http://www.xbrl.org/2003/instance"
    xmlns:link="http://www.xbrl.org/2003/linkbase"
    xmlns:xlink="http://www.w3.org/1999/xlink">
    <link:schemaRef xlink:type="simple" xlink:href="example.xsd"/>
    <link:linkbaseRef xlink:type="simple" xlink:href="labels.xml"/>
    <link:roleRef roleURI="http://example.com/role" xlink:href="roles.xml#role"/>
    <link:arcroleRef arcroleURI="http://example.com/arcrole" xlink:href="arcs.xml#arc"/>
    <link:footnoteLink>
        <link:footnote xlink:label="f1">Preserved footnote text</link:footnote>
    </link:footnoteLink>
    <item id="baseItem"/>
    <tuple id="baseTuple"><child>nested</child></tuple>
</xbrl>`

	var raw RawXBRL
	require.NoError(t, xml.Unmarshal([]byte(doc), &raw))

	require.Len(t, raw.SchemaRefs, 1)
	assert.Equal(t, "example.xsd", requireAttr(t, raw.SchemaRefs[0].Attributes, "http://www.w3.org/1999/xlink", "href"))
	assert.Len(t, raw.SchemaRef, 1)

	require.Len(t, raw.LinkbaseRefs, 1)
	assert.Equal(t, "labels.xml", requireAttr(t, raw.LinkbaseRefs[0].Attributes, "http://www.w3.org/1999/xlink", "href"))
	assert.Len(t, raw.LinkbaseRef, 1)

	require.Len(t, raw.RoleRefs, 1)
	assert.Equal(t, "http://example.com/role", requireAttr(t, raw.RoleRefs[0].Attributes, "", "roleURI"))
	assert.Len(t, raw.RoleRef, 1)

	require.Len(t, raw.ArcRoleRefs, 1)
	assert.Equal(t, "http://example.com/arcrole", requireAttr(t, raw.ArcRoleRefs[0].Attributes, "", "arcroleURI"))
	assert.Len(t, raw.ArcRoleRef, 1)

	require.Len(t, raw.FootnoteLinks, 1)
	assert.Contains(t, raw.FootnoteLinks[0].InnerXML, "Preserved footnote text")
	assert.Len(t, raw.FootnoteLink, 1)

	require.Len(t, raw.UnsupportedTopLevel, 2)
	assert.Equal(t, xml.Name{Space: xbrlInstanceNamespace, Local: "item"}, raw.UnsupportedTopLevel[0].XMLName)
	assert.Equal(t, xml.Name{Space: xbrlInstanceNamespace, Local: "tuple"}, raw.UnsupportedTopLevel[1].XMLName)
	assert.Contains(t, raw.UnsupportedTopLevel[1].InnerXML, "<child>nested</child>")
	assert.Empty(t, raw.Facts)
}

func TestValidateRejectsDuplicateContextIDs(t *testing.T) {
	// language=xml
	doc := `<xbrl>
    <link:schemaRef/>
    <context id="c1">
        <entity><identifier scheme="http://www.sec.gov/CIK">0000320193</identifier></entity>
        <period><instant>2021-03-27</instant></period>
    </context>
    <context id="c1">
        <entity><identifier scheme="http://www.sec.gov/CIK">0000320193</identifier></entity>
        <period><instant>2021-03-28</instant></period>
    </context>
    <unit id="u1"><measure>shares</measure></unit>
    <ci:assets contextRef="c1" unitRef="u1" precision="3">727</ci:assets>
</xbrl>`

	var content XBRL
	require.NoError(t, xml.Unmarshal([]byte(doc), &content))

	assert.EqualError(t, content.Validate(), "duplicate context id: c1")
}

func TestValidateRejectsDuplicateUnitIDs(t *testing.T) {
	// language=xml
	doc := `<xbrl>
    <link:schemaRef/>
    <context id="c1">
        <entity><identifier scheme="http://www.sec.gov/CIK">0000320193</identifier></entity>
        <period><instant>2021-03-27</instant></period>
    </context>
    <unit id="u1"><measure>shares</measure></unit>
    <unit id="u1"><measure>iso4217:USD</measure></unit>
    <ci:assets contextRef="c1" unitRef="u1" precision="3">727</ci:assets>
</xbrl>`

	var content XBRL
	require.NoError(t, xml.Unmarshal([]byte(doc), &content))

	assert.EqualError(t, content.Validate(), "duplicate unit id: u1")
}

func TestValidateRejectsKnownUnsupportedTopLevelElements(t *testing.T) {
	// language=xml
	doc := `<xbrl xmlns="http://www.xbrl.org/2003/instance">
    <tuple/>
</xbrl>`

	var content XBRL
	require.NoError(t, xml.Unmarshal([]byte(doc), &content))

	require.Len(t, content.UnsupportedTopLevel, 1)
	assert.EqualError(t, content.Validate(), "unsupported top-level element: http://www.xbrl.org/2003/instance:tuple")
}

func stringPtr(str string) *string {
	return &str
}

func requireAttr(t *testing.T, attrs []xml.Attr, space, local string) string {
	t.Helper()

	for _, attr := range attrs {
		if attr.Name.Space == space && attr.Name.Local == local {
			return attr.Value
		}
	}

	require.Failf(t, "missing XML attribute", "space=%q local=%q attrs=%v", space, local, attrs)
	return ""
}

func requireFact(t *testing.T, facts []Fact, space, local, value string) Fact {
	t.Helper()

	for _, fact := range facts {
		if fact.XMLName.Space == space && fact.XMLName.Local == local && fact.Value() == value {
			return fact
		}
	}

	require.Failf(t, "missing fact", "space=%q local=%q value=%q", space, local, value)
	return Fact{}
}
