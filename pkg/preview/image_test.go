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
			DataLength: 11312,
			Width:      750,
			Height:     1000,
			Hash:       "a7dfa90ec80959dd4239a48b394c5e0e98aaabc57710f8691bf0244f6d15ce23",
		},
		{
			Name:       ".gif",
			SourcePath: "../../test/sample_minions.gif",
			DataLength: 8970,
			Width:      400,
			Height:     200,
			Hash:       "118fbbdad8bfe93ac4982dfac7f26e418b5354eece720a78359e3ca1bc209757",
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
