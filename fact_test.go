package xbrl

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalFact(t *testing.T) {
	t.Run("nil fact", func(t *testing.T) {
		// language=xml
		factXML := `<myns:sillyFact contextRef="c1" xsi:nil="true"/>`

		var fact Fact
		require.NoError(t, xml.Unmarshal([]byte(factXML), &fact))

		assert.Equal(t, xml.Name{Space: "myns", Local: "sillyFact"}, fact.XMLName)
		assert.Equal(t, FactTypeNil, fact.Type())
		assert.True(t, fact.IsValid())
		assert.Equal(t, "c1", fact.ContextRef)
	})

	t.Run("non-numeric fact", func(t *testing.T) {
		// language=xml
		factXML := `<ci:concentrationsNote contextRef="c1">Some cool text block about concentrations.</ci:concentrationsNote>`

		var fact Fact
		require.NoError(t, xml.Unmarshal([]byte(factXML), &fact))

		assert.Equal(t, xml.Name{Space: "ci", Local: "concentrationsNote"}, fact.XMLName)
		assert.Equal(t, FactTypeNonNumeric, fact.Type())
		assert.True(t, fact.IsValid())
		assert.Equal(t, "c1", fact.ContextRef)
		assert.Nil(t, fact.UnitRef)
		assert.Equal(t, "Some cool text block about concentrations.", fact.Value())
	})

	t.Run("simple numeric fact", func(t *testing.T) {
		// language=xml
		factXML := `<ci:capitalLeases id="id123" contextRef="c1" unitRef="u1" precision="3">727432</ci:capitalLeases>`

		var fact Fact
		require.NoError(t, xml.Unmarshal([]byte(factXML), &fact))

		assert.Equal(t, xml.Name{Space: "ci", Local: "capitalLeases"}, fact.XMLName)
		assert.Equal(t, FactTypeNonFraction, fact.Type())
		assert.True(t, fact.IsValid())
		assert.Equal(t, "id123", fact.ID)
		assert.Equal(t, "c1", fact.ContextRef)
		require.NotNil(t, fact.UnitRef)
		assert.Equal(t, "u1", *fact.UnitRef)
		require.NotNil(t, fact.Precision)
		assert.Equal(t, "3", *fact.Precision)
		assert.Nil(t, fact.Decimals)

		val, err := fact.NumericValue()
		require.NoError(t, err)
		assert.EqualValues(t, 727432, val)
	})

	t.Run("simple decimal numeric fact", func(t *testing.T) {
		// language=xml
		factXML := `<us-gaap:EarningsPerShareBasic contextRef="i0ad" decimals="2" id="id3Vyb" unitRef="usdPerShare">0.64</us-gaap:EarningsPerShareBasic>`

		var fact Fact
		require.NoError(t, xml.Unmarshal([]byte(factXML), &fact))

		assert.Equal(t, xml.Name{Space: "us-gaap", Local: "EarningsPerShareBasic"}, fact.XMLName)
		assert.Equal(t, FactTypeNonFraction, fact.Type())
		assert.True(t, fact.IsValid())
		assert.Equal(t, "id3Vyb", fact.ID)
		assert.Equal(t, "i0ad", fact.ContextRef)
		require.NotNil(t, fact.UnitRef)
		assert.Equal(t, "usdPerShare", *fact.UnitRef)
		assert.Nil(t, fact.Precision)
		require.NotNil(t, fact.Decimals)
		assert.Equal(t, "2", *fact.Decimals)

		val, err := fact.NumericValue()
		require.NoError(t, err)
		assert.EqualValues(t, 0.64, val)
	})

	t.Run("fraction type numeric fact", func(t *testing.T) {
		// language=xml
		factXML := `<myTaxonomy:oneThird id="oneThird" unitRef="u1" contextRef="numC1">
	<numerator>1</numerator>
	<denominator>3</denominator>
</myTaxonomy:oneThird>`

		var fact Fact
		require.NoError(t, xml.Unmarshal([]byte(factXML), &fact))

		assert.Equal(t, xml.Name{Space: "myTaxonomy", Local: "oneThird"}, fact.XMLName)
		assert.Equal(t, FactTypeFraction, fact.Type())
		assert.True(t, fact.IsValid())
		assert.Equal(t, "oneThird", fact.ID)
		assert.Equal(t, "numC1", fact.ContextRef)
		require.NotNil(t, fact.UnitRef)
		assert.Equal(t, "u1", *fact.UnitRef)
		assert.Nil(t, fact.Precision)
		assert.Nil(t, fact.Decimals)

		val, err := fact.NumericValue()
		require.NoError(t, err)
		assert.EqualValues(t, 1.0/3.0, val)
	})
}

func TestFactValidation(t *testing.T) {
	unitRef := "u1"
	precision := "3"
	decimals := "2"
	invalidPrecision := "-1"
	invalidDecimals := "not-an-integer"

	tests := []struct {
		name    string
		fact    Fact
		wantErr string
	}{
		{
			name: "fact requires context ref",
			fact: Fact{
				UnitRef:   &unitRef,
				Precision: &precision,
				ValueStr:  stringPtr("727"),
			},
			wantErr: "missing contextRef",
		},
		{
			name: "non-fraction missing value",
			fact: Fact{
				ContextRef: "c1",
				UnitRef:    &unitRef,
				Precision:  &precision,
			},
			wantErr: "non-fraction fact missing value",
		},
		{
			name: "non-fraction with precision and decimals",
			fact: Fact{
				ContextRef: "c1",
				UnitRef:    &unitRef,
				Precision:  &precision,
				Decimals:   &decimals,
				ValueStr:   stringPtr("727"),
			},
			wantErr: "non-fraction fact must have exactly one of precision or decimals",
		},
		{
			name: "non-fraction with invalid precision",
			fact: Fact{
				ContextRef: "c1",
				UnitRef:    &unitRef,
				Precision:  &invalidPrecision,
				ValueStr:   stringPtr("727"),
			},
			wantErr: "non-fraction fact has invalid precision",
		},
		{
			name: "non-fraction with invalid decimals",
			fact: Fact{
				ContextRef: "c1",
				UnitRef:    &unitRef,
				Decimals:   &invalidDecimals,
				ValueStr:   stringPtr("727"),
			},
			wantErr: "non-fraction fact has invalid decimals",
		},
		{
			name: "non-numeric with precision",
			fact: Fact{
				ContextRef: "c1",
				Precision:  &precision,
				ValueStr:   stringPtr("not numeric"),
			},
			wantErr: "non-numeric fact cannot have precision or decimals",
		},
		{
			name: "fraction with zero denominator",
			fact: Fact{
				ContextRef:  "c1",
				UnitRef:     &unitRef,
				Numerator:   floatPtr(1),
				Denominator: floatPtr(0),
			},
			wantErr: "fraction fact denominator is zero",
		},
		{
			name: "fraction with precision",
			fact: Fact{
				ContextRef:  "c1",
				UnitRef:     &unitRef,
				Precision:   &precision,
				Numerator:   floatPtr(1),
				Denominator: floatPtr(3),
			},
			wantErr: "fraction fact cannot have precision or decimals",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fact.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.False(t, tt.fact.IsValid())
		})
	}
}

func TestNilFactValidation(t *testing.T) {
	nilValue := true
	fact := Fact{
		XMLName:    xml.Name{Space: "myns", Local: "nilFact"},
		Nil:        &nilValue,
		ContextRef: "c1",
	}

	assert.Equal(t, FactTypeNil, fact.Type())
	assert.NoError(t, fact.Validate())
}

func TestNumericValueMalformedFactReturnsError(t *testing.T) {
	unitRef := "u1"
	precision := "3"
	fact := Fact{
		ContextRef: "c1",
		UnitRef:    &unitRef,
		Precision:  &precision,
	}

	_, err := fact.NumericValue()
	assert.Error(t, err)
}

func floatPtr(val float64) *float64 {
	return &val
}
