package xbrl

import (
	"encoding/xml"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHelpers(t *testing.T) {
	t.Run("Parse does not validate", func(t *testing.T) {
		xbrlBytes, err := os.ReadFile("test_data/invalid_xbrl.xml")
		require.NoError(t, err)

		doc, err := Parse(xbrlBytes)
		require.NoError(t, err)

		assert.Error(t, doc.Validate())
	})

	t.Run("ParseReader", func(t *testing.T) {
		doc, err := ParseReader(strings.NewReader(`<xbrl/>`))
		require.NoError(t, err)

		assert.Equal(t, xml.Name{Local: "xbrl"}, doc.XMLName)
	})

	t.Run("Decode uses caller configured decoder", func(t *testing.T) {
		f, err := os.Open("test_data/edgr-2004_10k.xml")
		require.NoError(t, err)
		defer f.Close()

		decoder := xml.NewDecoder(f)
		decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
			return input, nil
		}

		doc, err := Decode(decoder)
		require.NoError(t, err)

		assert.Equal(t, xml.Name{Space: xbrlInstanceNamespace, Local: "xbrl"}, doc.XMLName)
		assert.Len(t, doc.ContextsByID, 4)
	})

	t.Run("nil inputs", func(t *testing.T) {
		_, err := ParseReader(nil)
		assert.EqualError(t, err, "nil reader")

		_, err = Decode(nil)
		assert.EqualError(t, err, "nil decoder")
	})
}
