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
	video bool
	image bool
	path  string
	size  int64
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
			dataHash: "64de9c944a91c93e750d097577c8fc5992100a7bb186d376534e78705aefbbbd",
		},
		{
			source:   mockSource{video: true, path: "../../test/sample.mp4"},
			dataHash: "7f2d1e244e636db9178a8cb49d375ce3268048fddd9efa1d7943e03003a2b708",
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
