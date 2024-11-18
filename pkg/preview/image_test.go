package preview

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewImage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name       string
		SourcePath string
		DataLength int
		Width      int
		Height     int
		Hash       string
	}{
		{
			Name:       ".jpg",
			SourcePath: "../../test/image.jpg",
			DataLength: 3496,
			Width:      1527,
			Height:     898,
			Hash:       "0c8eaa3df6f838cbb5ece3ad12d1556fb781c2556dbe007c7b74b4e1df8c49f8",
		},
		{
			Name:       ".gif",
			SourcePath: "../../test/sample_minions.gif",
			DataLength: 7582,
			Width:      400,
			Height:     200,
			Hash:       "ac8e6312e7820089127bce5ca94242a579283428f70ec95b0777f5c6c1dba9f1",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)
			preview, err := imagePreview(testCase.SourcePath)
			require.NoError(t, err)
			assert.Len(preview.Data(), testCase.DataLength)
			width, height := preview.Resolution()
			assert.Equal(testCase.Width, width)
			assert.Equal(testCase.Height, height)

			hash := sha256.Sum256(preview.Data())
			assert.Equal(testCase.Hash, hex.EncodeToString(hash[:]))
		})
	}
}
