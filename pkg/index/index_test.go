package index

import (
	"bufio"
	"bytes"
	"context"
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
	t.Parallel()

	indexOriginal, err := NewIndex(context.Background(), nil)
	require.NoError(t, err)

	indexOriginal.meta = map[string]*Meta{
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
			Preview: Preview{
				Length: 100,
				Offset: 0,
			},
		},
	}
	indexOriginal.data = make([]byte, 100)
	_, err = rand.Read(indexOriginal.data)
	require.NoError(t, err)

	buff := new(bytes.Buffer)
	wrote, err := indexOriginal.Encode(buff)
	require.NoError(t, err)
	assert.EqualValues(t, 466, wrote)
	assert.EqualValues(t, 466, buff.Len())
	// Parse
	indexDerivative, err := NewIndex(context.Background(), bufio.NewReader(buff))
	require.NoError(t, err)
	assert.Len(t, indexDerivative.meta, len(indexOriginal.meta))
	assert.Equal(t, indexOriginal.meta, indexDerivative.meta)
	assert.Len(t, indexDerivative.data, len(indexOriginal.data))
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

func (mock mockPreviewGenerator) Pull(_ string) (preview.Data, error) {
	return preview.Data{Duration: 0, ContentType: ContentTypeOther, Resolution: "", Data: mock.sampleData}, nil
}

func (mock mockPreviewGenerator) ContentType(_ string) int {
	return ContentTypeOther
}

func TestIndexFiles(t *testing.T) {
	t.Parallel()

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

		file := &mockFile{stat, path, relativePath}
		filesMeta = append(filesMeta, file)
	}

	sampleData := make([]byte, 1000)
	index, err := NewIndex(
		context.Background(),
		nil,
		WithFiles(filesMeta),
		WithPreview(mockPreviewGenerator{sampleData: sampleData}),
		WithID(func(p FileMeta) (string, error) {
			return fmt.Sprintf("%s%s", p.RelativePath(), p.ModTime()), nil
		}),
	)
	require.NoError(t, err)
	assert.Len(t, index.meta, 7)
	assert.Len(t, index.data, len(filesNames)*len(sampleData))
}
