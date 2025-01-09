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
			DataLength: 15260,
			Width:      256,
			Height:     341,
			Hash:       "916594618fa51ee0084de8525d867a53e7e9f2a9b8544d8c14eb2c17a45c7632",
		},
		{
			Name:       ".gif",
			SourcePath: "../../test/sample_minions.gif",
			DataLength: 18700,
			Width:      512,
			Height:     256,
			Hash:       "262882df79b599b6a719f9a5f9dc0ed7ca541ca6897c290f762cd795e6edcec8",
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
