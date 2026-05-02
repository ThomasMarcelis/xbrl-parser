package xbrl

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalUnit(t *testing.T) {
	t.Run("simple unit", func(t *testing.T) {
		// language=xml
		unitXML := `<unit>
			<measure>shares</measure>
		</unit>`

		var unit Unit
		require.NoError(t, xml.Unmarshal([]byte(unitXML), &unit))

		require.Len(t, unit.Measures, 1)
		assert.Equal(t, "shares", unit.Measures[0].Value)
		assert.Equal(t, "shares", unit.Measures[0].String())
		assert.Nil(t, unit.Divide)

		assert.Equal(t, unit.String(), "shares")
	})

	t.Run("product of measures", func(t *testing.T) {
		// language=xml
		unitXML := `<unit>
			<measure>myns:feet</measure>
			<measure>myns:feet</measure>
		</unit>`

		var unit Unit
		require.NoError(t, xml.Unmarshal([]byte(unitXML), &unit))

		require.Len(t, unit.Measures, 2)
		assert.Equal(t, "myns:feet", unit.Measures[0].Value)
		assert.Equal(t, "feet", unit.Measures[0].String())

		assert.Equal(t, "myns:feet", unit.Measures[1].Value)
		assert.Equal(t, "feet", unit.Measures[1].String())

		assert.Nil(t, unit.Divide)

		assert.Equal(t, "feet * feet", unit.String())
	})

	t.Run("ratio of simple measures", func(t *testing.T) {
		// language=xml
		unitXML := `<unit>
			<divide>
				<unitNumerator>
					<measure>iso4127:USD</measure>
				</unitNumerator>
				<unitDenominator>
					<measure>shares</measure>
				</unitDenominator>
			</divide>
		</unit>`

		var unit Unit
		require.NoError(t, xml.Unmarshal([]byte(unitXML), &unit))

		require.Len(t, unit.Measures, 0)

		assert.NotNil(t, unit.Divide)
		assert.Len(t, unit.Divide.Numerator, 1)
		assert.Equal(t, "iso4127:USD", unit.Divide.Numerator[0].Value)

		assert.Len(t, unit.Divide.Denominator, 1)
		assert.Equal(t, "shares", unit.Divide.Denominator[0].Value)

		assert.Equal(t, "USD / shares", unit.String())
	})

	t.Run("ratio of products of measures", func(t *testing.T) {
		// language=xml
		unitXML := `<unit>
			<divide>
				<unitNumerator>
					<measure>iso4127:USD</measure>
				</unitNumerator>
				<unitDenominator>
					<measure>myns:feet</measure>
					<measure>myns:feet</measure>
				</unitDenominator>
			</divide>
		</unit>`

		var unit Unit
		require.NoError(t, xml.Unmarshal([]byte(unitXML), &unit))

		require.Len(t, unit.Measures, 0)

		assert.NotNil(t, unit.Divide)
		assert.Len(t, unit.Divide.Numerator, 1)
		assert.Equal(t, "iso4127:USD", unit.Divide.Numerator[0].Value)

		assert.Len(t, unit.Divide.Denominator, 2)
		assert.Equal(t, "myns:feet", unit.Divide.Denominator[0].Value)
		assert.Equal(t, "myns:feet", unit.Divide.Denominator[1].Value)

		assert.Equal(t, "USD / feet * feet", unit.String())
	})
}

func TestUnitValidation(t *testing.T) {
	tests := []struct {
		name    string
		unit    Unit
		wantErr string
	}{
		{
			name:    "unit requires id",
			unit:    Unit{Measures: Measures{{Value: "shares"}}},
			wantErr: "unit missing id",
		},
		{
			name:    "unit requires measure or divide",
			unit:    Unit{ID: "u1"},
			wantErr: "unit must have either measures or divide",
		},
		{
			name: "unit cannot have measures and divide",
			unit: Unit{
				ID:       "u1",
				Measures: Measures{{Value: "shares"}},
				Divide: &Divide{
					Numerator:   Measures{{Value: "iso4217:USD"}},
					Denominator: Measures{{Value: "shares"}},
				},
			},
			wantErr: "unit must have either measures or divide",
		},
		{
			name: "divide requires denominator measures",
			unit: Unit{
				ID: "u1",
				Divide: &Divide{
					Numerator: Measures{{Value: "iso4217:USD"}},
				},
			},
			wantErr: "divide missing denominator measures",
		},
		{
			name: "measure requires value",
			unit: Unit{
				ID:       "u1",
				Measures: Measures{{}},
			},
			wantErr: "measure missing value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.unit.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.False(t, tt.unit.IsValid())
		})
	}
}

func TestMeasureValuePreservesRawQName(t *testing.T) {
	measure := Measure{Value: "iso4217:USD"}

	assert.Equal(t, "iso4217:USD", measure.Value)
	assert.Equal(t, "USD", measure.String())
}
