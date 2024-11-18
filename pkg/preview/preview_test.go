package preview

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSource struct {
	video     bool
	image     bool
	path      string
	extension string
	size      int64
}

func (s mockSource) IsImage() bool {
	return s.image
}

func (s mockSource) IsVideo() bool {
	return s.video
}

func (s mockSource) Path() string {
	return s.path
}

func (s mockSource) Size() int64 {
	return s.size
}

func (s mockSource) IsAnimatedImage() bool {
	return s.extension == "gif"
}

func TestPreview(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	previewer, err := NewPreviewer(WithImage(true), WithVideo(true))
	require.NoError(err)

	testCases := []struct {
		source   mockSource
		dataHash string
	}{
		{
			source:   mockSource{image: true, path: "../../test/image.jpg"},
			dataHash: "0c8eaa3df6f838cbb5ece3ad12d1556fb781c2556dbe007c7b74b4e1df8c49f8",
		},
		{
			source:   mockSource{video: true, path: "../../test/sample.mp4"},
			dataHash: "df04235b400d9e0d623681102ca40c7424ca5efff0014e4e1b74d0452c9030fb",
		},
		{
			source:   mockSource{path: "../../test/sample.txt"},
			dataHash: "",
		},
	}

	for _, tCase := range testCases {
		t.Run(filepath.Ext(tCase.source.Path()), func(t *testing.T) {
			t.Parallel()

			preview, err := previewer.Pull(tCase.source)
			require.NoError(err)
			assert.Equal(t, preview.Data() == nil, tCase.dataHash == "")

			if tCase.dataHash == "" {
				return
			}

			hash := sha256.Sum256(preview.Data())
			assert.Equal(t, tCase.dataHash, hex.EncodeToString(hash[:]))
		})
	}
}
