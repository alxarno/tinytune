package index

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alxarno/tinytune/pkg/preview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexEncodeDecode(t *testing.T) {
	indexOriginal, err := NewIndex(nil)
	require.NoError(t, err)
	indexOriginal.meta = map[string]*IndexMeta{
		"5762029e772": {
			Path:         "/home/test/",
			Name:         "test",
			ID:           "5762029e772",
			RelativePath: "test/",
			ModTime:      time.Date(2024, 11, 5, 5, 5, 5, 0, time.UTC),
			Type:         ContentTypeOther,
			IsDir:        true,
		},
		"2cf24dba5fb": {
			Path:         "/home/test/abc.jpg",
			Name:         "abc.jpg",
			ID:           "2cf24dba5fb",
			RelativePath: "test/abc.jpg",
			ModTime:      time.Date(2024, 11, 5, 5, 5, 5, 0, time.UTC),
			Type:         ContentTypeImage,
			IsDir:        false,
			Preview: IndexMetaPreview{
				Length: 100,
				Offset: 0,
			},
		},
	}
	indexOriginal.data = make([]byte, 100)
	rand.Read(indexOriginal.data)
	buff := new(bytes.Buffer)
	wrote, err := indexOriginal.Encode(buff)
	assert.NoError(t, err)
	assert.EqualValues(t, 480, wrote)
	assert.EqualValues(t, 480, buff.Len())
	// Parse
	indexDerivative, err := NewIndex(bufio.NewReader(buff))
	require.NoError(t, err)
	assert.Equal(t, len(indexOriginal.meta), len(indexDerivative.meta))
	assert.Equal(t, indexOriginal.meta, indexDerivative.meta)
	assert.Equal(t, len(indexOriginal.data), len(indexDerivative.data))
	assert.Equal(t, indexOriginal.data, indexDerivative.data)
}

type mockFile struct {
	os.FileInfo
	path         string
	relativePath string
}

func (mock mockFile) Path() string {
	return mock.path
}

func (mock mockFile) RelativePath() string {
	return mock.relativePath
}

type mockPreviewGenerator struct {
	sampleData []byte
}

func (mock mockPreviewGenerator) Pull(path string) (preview.PreviewData, error) {
	return preview.PreviewData{Duration: 0, ContentType: ContentTypeOther, Resolution: "", Data: mock.sampleData}, nil
}

func (mock mockPreviewGenerator) ContentType(path string) int {
	return ContentTypeOther
}

func TestIndexFiles(t *testing.T) {
	testFolderPath, _ := strings.CutSuffix(os.Getenv("PWD"), "pkg/index")
	testFolderPath = filepath.Join(testFolderPath, "test")
	filesNames := []string{
		filepath.Join(testFolderPath, "image.jpg"),
		filepath.Join(testFolderPath, "sample.mp4"),
		filepath.Join(testFolderPath, "sample.txt"),
		filepath.Join(testFolderPath, "video"),
		filepath.Join(testFolderPath, "video/sample.mp4"),
		filepath.Join(testFolderPath, "img"),
		filepath.Join(testFolderPath, "img/image.jpg"),
	}
	filesMeta := []FileMeta{}
	for _, path := range filesNames {
		stat, err := os.Stat(path)
		require.NoError(t, err)
		relativePath, err := filepath.Rel(testFolderPath, path)
		require.NoError(t, err)
		filesMeta = append(filesMeta, &mockFile{stat, path, relativePath})
	}

	sampleData := make([]byte, 1000)
	index, err := NewIndex(
		nil,
		WithFiles(filesMeta),
		WithPreview(mockPreviewGenerator{sampleData: sampleData}),
		WithID(func(p FileMeta) (string, error) {
			return fmt.Sprintf("%s%s", p.RelativePath(), p.ModTime()), nil
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, 7, len(index.meta))
	assert.EqualValues(t, len(filesNames)*len(sampleData), len(index.data))
}
