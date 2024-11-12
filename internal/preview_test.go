package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreview(t *testing.T) {
	t.Parallel()

	previewer, err := NewPreviewer(WithImagePreview(true), WithVideoPreview(true))
	require.NoError(t, err)

	testCases := []struct {
		path        string
		hash        string
		resultIsNil bool
	}{
		{"../test/image.jpg", "64de9c944a91c93e750d097577c8fc5992100a7bb186d376534e78705aefbbbd", false},
		{"../test/sample.mp4", "913e1f20eb400f3a13aa043005204ef53e0883c122086b96d94a2b6279ec008e", false},
		{"../test/sample.txt", "", true},
	}

	for _, tCase := range testCases {
		t.Run(filepath.Ext(tCase.path), func(t *testing.T) {
			t.Parallel()

			preview, err := previewer.Pull(tCase.path)
			require.NoError(t, err)
			assert.Equal(t, preview.Data == nil, tCase.resultIsNil)

			if tCase.resultIsNil {
				return
			}

			hash := sha256.Sum256(preview.Data)
			assert.Equal(t, tCase.hash, hex.EncodeToString(hash[:]))
		})
	}
}
