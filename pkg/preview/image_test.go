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
			DataLength: 10332,
			Width:      750,
			Height:     1000,
			Hash:       "182e0008de9bdd21a8af91ce6e6cbc57197e604886de71791133147f217c6fa0",
		},
		{
			Name:       ".gif",
			SourcePath: "../../test/sample_minions.gif",
			DataLength: 7794,
			Width:      400,
			Height:     200,
			Hash:       "846c829de9f1e4490e7025dcba99c903a7cb93289cd669b28790ffbc838e3847",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)
			preview, err := imagePreview(testCase.SourcePath, 1<<10)
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
