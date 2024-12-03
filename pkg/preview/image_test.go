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
			SourcePath: "../../test/img/image.jpg",
			DataLength: 9946,
			Width:      750,
			Height:     1000,
			Hash:       "08b43a0683e84c4b26f61c2e143f813e11f08808a792ab86d819c69434db715a",
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
