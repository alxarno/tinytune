package internal

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreview(t *testing.T) {
	testCases := []struct {
		path        string
		resultHash  string
		resultIsNil bool
	}{
		{"../test/image.jpg", "64de9c944a91c93e750d097577c8fc5992100a7bb186d376534e78705aefbbbd", false},
		{"../test/sample.mp4", "913e1f20eb400f3a13aa043005204ef53e0883c122086b96d94a2b6279ec008e", false},
		{"../test/sample.txt", "", true},
	}
	for _, tc := range testCases {
		previewer, err := NewPreviewer(WithImagePreview(), WithVideoPreview())
		require.NoError(t, err)
		t.Run(filepath.Ext(tc.path), func(tt *testing.T) {
			preview, err := previewer.Pull(tc.path)
			assert.NoError(tt, err)
			assert.Equal(tt, preview.Data == nil, tc.resultIsNil)
			if tc.resultIsNil {
				return
			}
			hash, err := SHA256Hash(bytes.NewReader(preview.Data))
			assert.NoError(tt, err)
			assert.Equal(tt, tc.resultHash, hash)
		})
	}
}
