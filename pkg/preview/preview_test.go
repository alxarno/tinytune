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
			source:   mockSource{image: true, path: "../../test/img/image.jpg"},
			dataHash: "08b43a0683e84c4b26f61c2e143f813e11f08808a792ab86d819c69434db715a",
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
