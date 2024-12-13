package index

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIndexBuilder(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	indexFile, err := os.Open("../../test/test.index.tinytune")
	require.NoError(err)
	index, err := NewIndex(context.Background(), indexFile)
	require.NoError(err)
	require.Len(index.meta, 17)
	sample, ok := index.meta["623f14247e"]
	require.True(ok)
	require.True(sample.IsVideo())
	require.Equal("sample_960x400_ocean_with_audio.flv", sample.Name)
	require.Equal(960, sample.Resolution.Width)
	require.Equal(400, sample.Resolution.Height)
	require.EqualValues(12534, sample.Preview.Length)
	require.EqualValues(36234, sample.Preview.Offset)
	require.EqualValues(7222114, sample.OriginSize)
	require.LessOrEqual(sample.Preview.Offset+sample.Preview.Length, uint32(len(index.data)))
}

func TestIndexBuilderClearRemovedFiles(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	testFolderPath, _ := strings.CutSuffix(os.Getenv("PWD"), "pkg/index")
	testFolderPath = filepath.Join(testFolderPath, "test")
	indexFile, err := os.Open(filepath.Join(testFolderPath, "test.index.tinytune"))
	require.NoError(err)
	require.NotNil(indexFile)

	dataPartLength := 172688
	removedPreviewLength := 7582
	// nano timestamps from test.index.tinytune
	filesPaths := []struct {
		path    string
		modTime int64
	}{
		// removed item
		// {"sample_minions.gif", 1731931322},
		{"test.m3u8", 1733862508},
		{"sample.txt", 1730413055},
		{"2.ts", 1731604682},
		{"video/sample_960x400_ocean_with_audio.flv", 1731524953},
		{"image.jpg", 1734099668},
		{"img/nested/nested-image.jpg", 1730976772},
		{"img", 1730976766},
		{"video/sample.mp4", 1730213589},
		{"video", 1731524985},
		{"Anh_nude_cover.webp", 1731091398},
		{"test.index.tinytune", 1734111561},
		{"img/nested", 1730976778},
		{"long-name-sample-for-testing-names-displaying-1111122223333.txt", 1731060108},
		{"img/image.jpg", 1730213353},
		{"short.mp4", 1731872015},
		{"sample.mp4", 1730213480},
	}

	filesMeta := []FileMeta{}

	for _, pathItem := range filesPaths {
		path := filepath.Join(testFolderPath, pathItem.path)
		stat, err := os.Stat(path)
		require.NoError(err)
		relativePath, err := filepath.Rel(testFolderPath, path)
		require.NoError(err)

		file := &mockFile{
			FileInfo:     stat,
			path:         path,
			relativePath: relativePath,
			dir:          stat.IsDir(),
			// index will be generate new id, if mod stat diff in index file and FS
			modTime: time.Unix(pathItem.modTime, 0),
		}
		filesMeta = append(filesMeta, file)
	}

	index, err := NewIndex(
		context.Background(),
		indexFile,
		WithFiles(filesMeta),
		WithRemovedFilesCleaning(),
	)
	require.NoError(err)
	require.Len(index.meta, len(filesPaths))

	// image.jpg file has 66002 (offset) 8620 (length)
	// will test preview
	var image *Meta

	for _, v := range index.meta {
		if v.RelativePath == RelativePath("image.jpg") {
			image = v
		}
	}

	require.NotNil(image)
	require.NotEqual(0, image.Preview.Length)
	require.NotEqual(0, image.Preview.Offset)
	require.Len(index.data, dataPartLength-removedPreviewLength)

	m, err := index.Pull(image.ID)
	require.NoError(err)
	require.NotNil(m)

	itemNextPreviewDataHash := "b9237e340eb886b5e437e4bc6fd19be612ff6653f3f733c3998ef20f320e6c20"

	imagePrevew, err := index.PullPreview(image.ID)
	require.NoError(err)
	require.NotNil(imagePrevew)

	h := sha256.New()
	h.Write(imagePrevew)

	require.Equal(itemNextPreviewDataHash, hex.EncodeToString(h.Sum(nil)))
}
