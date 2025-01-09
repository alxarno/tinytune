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
			dataHash: "916594618fa51ee0084de8525d867a53e7e9f2a9b8544d8c14eb2c17a45c7632",
		},
		{
			source:   mockSource{video: true, path: "../../test/sample.mp4"},
			dataHash: "b5039ea0fe138cd39b7c5b1b1bad83324fcb02f73862094a3408bc74ea522b36",
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
