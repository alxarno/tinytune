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
			DataLength: 8620,
			Width:      1527,
			Height:     898,
			Hash:       "64de9c944a91c93e750d097577c8fc5992100a7bb186d376534e78705aefbbbd",
		},
		{
			Name:       ".gif",
			SourcePath: "../../test/sample_minions.gif",
			DataLength: 15322,
			Width:      400,
			Height:     200,
			Hash:       "3639b076b0df5d917bd6f439aa8f11a3fc2cdcc638105a5a52e670d2c3105e3a",
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
