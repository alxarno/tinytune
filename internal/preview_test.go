package internal

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreview(t *testing.T) {
	testCases := []struct {
		path        string
		resultHash  string
		resultIsNil bool
	}{
		{"../test/image.jpg", "029f21ed0e085973dc41f290f5d361185f226c924a7e30e2d2d8c8acac6ade5a", false},
		{"../test/sample.mp4", "b29ec13ece50f8343604a26f83216153f3ea58f038d00e1fae9e4461c6c36313", false},
		{"../test/sample.txt", "", true},
	}
	for _, tc := range testCases {
		t.Run(filepath.Ext(tc.path), func(tt *testing.T) {
			_, _, preview, err := GeneratePreview(tc.path)
			assert.NoError(tt, err)
			assert.Equal(tt, preview == nil, tc.resultIsNil)
			if tc.resultIsNil {
				return
			}
			hash, err := SHA256Hash(bytes.NewReader(preview))
			assert.NoError(tt, err)
			assert.Equal(tt, tc.resultHash, hash)
		})
	}
}
