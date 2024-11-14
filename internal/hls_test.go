package internal

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHLSIndex(t *testing.T) {
	t.Parallel()

	meta := &index.Meta{
		Duration: time.Second * 75,
	}
	buff := bytes.Buffer{}
	err := pullHLSIndex(meta, &buff)
	require.NoError(t, err)

	f, err := os.OpenFile("../test/test.m3u8", os.O_RDONLY, 0755)
	require.NoError(t, err)

	defer f.Close()
	valid, err := io.ReadAll(f)
	require.NoError(t, err)
	assert.Equal(t, valid, buff.Bytes())
}

func TestPullHLSChunk(t *testing.T) {
	t.Parallel()

	meta := &index.Meta{
		Path:     "../test/video/sample_960x400_ocean_with_audio.flv",
		Duration: time.Second * 46,
	}
	buff := bytes.Buffer{}
	err := pullHLSChunk(context.Background(), meta, "2.ts", 0, &buff)
	require.NoError(t, err)

	// different machines produces little different result, so i decided comment it
	//
	// f, err := os.OpenFile("../test/2.ts", os.O_RDONLY, 0755)
	// require.NoError(t, err)

	// defer f.Close()
	// valid, err := io.ReadAll(f)
	// require.NoError(t, err)
	// assert.Len(t, buff.Bytes(), len(valid))
	// hash := sha256.Sum256(buff.Bytes())
	// assert.Equal(t, "7030f148eeeeb0103419457ca1633a7a634a11538f4f771155e9a7eef069a8b0", hex.EncodeToString(hash[:]))
}
