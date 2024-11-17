package index

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"io/fs"
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

	indexOriginal.meta = map[ID]*Meta{
		"5762029e772": {
			AbsolutePath: "/home/test/",
			Name:         "test",
			ID:           "5762029e772",
			RelativePath: "test/",
			ModTime:      time.Date(2024, 11, 5, 5, 5, 5, 0, time.UTC),
			Type:         ContentTypeOther,
			IsDir:        true,
		},
		"2cf24dba5fb": {
			AbsolutePath: "/home/test/abc.jpg",
			Name:         "abc.jpg",
			ID:           "2cf24dba5fb",
			RelativePath: "test/abc.jpg",
			ModTime:      time.Date(2024, 11, 5, 5, 5, 5, 0, time.UTC),
			Type:         ContentTypeImage,
			IsDir:        false,
			Preview: PreviewLocation{
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
	assert.NotEqualValues(t, 0, wrote)
	assert.NotEqualValues(t, 0, buff.Len())
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
	dir          bool
	path         string
	relativePath string
	modTime      time.Time
	mode         fs.FileMode
	name         string
	size         int64
}

func (mock mockFile) Path() string {
	return mock.path
}

func (mock mockFile) RelativePath() string {
	return mock.relativePath
}

func (mock mockFile) IsDir() bool {
	return mock.dir
}

func (mock mockFile) ModTime() time.Time {
	return mock.modTime
}
func (mock mockFile) Mode() fs.FileMode {
	return mock.mode
}
func (mock mockFile) Name() string {
	return mock.name
}
func (mock mockFile) Size() int64 {
	return mock.size
}

type mockPreviewGenerator struct {
	sampleData []byte
}

type mockPreviewData struct {
	data []byte
}

func (m mockPreviewData) Data() []byte {
	return m.data
}

func (m mockPreviewData) Resolution() (int, int) {
	return 0, 0
}

func (m mockPreviewData) Duration() time.Duration {
	return 0
}

//nolint:ireturn
func (mock mockPreviewGenerator) Pull(_ preview.Source) (preview.Data, error) {
	return mockPreviewData{data: mock.sampleData}, nil
}

func TestIndexFiles(t *testing.T) {
	t.Parallel()

	testFolderPath, _ := strings.CutSuffix(os.Getenv("PWD"), "pkg/index")
	testFolderPath = filepath.Join(testFolderPath, "test")
	filesWithPreview := 5
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

		file := &mockFile{
			FileInfo:     stat,
			path:         path,
			relativePath: relativePath,
			dir:          stat.IsDir(),
			modTime:      stat.ModTime(),
		}
		filesMeta = append(filesMeta, file)
	}

	sampleData := make([]byte, 1000)
	previewer := mockPreviewGenerator{sampleData: sampleData}
	index, err := NewIndex(
		context.Background(),
		nil,
		WithFiles(filesMeta),
		WithPreview(previewer),
	)
	require.NoError(t, err)
	assert.Len(t, index.meta, 7)
	assert.Len(t, index.data, filesWithPreview*len(sampleData))
}

func TestIndexUpdatedFile(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	indexOriginal, err := NewIndex(context.Background(), nil)
	require.NoError(err)

	indexOriginal.meta = map[ID]*Meta{
		"2cf24dba5fb": {
			AbsolutePath: "/home/test/abc.jpg",
			Name:         "abc.jpg",
			ID:           "2cf24dba5fb",
			RelativePath: "test/abc.jpg",
			ModTime:      time.Date(2024, 11, 5, 5, 5, 5, 0, time.UTC),
			Type:         ContentTypeImage,
			IsDir:        false,
			Preview: PreviewLocation{
				Length: 100,
				Offset: 0,
			},
		},
	}
	buff := new(bytes.Buffer)
	_, err = indexOriginal.Encode(buff)
	require.NoError(err)
	// Parse
	originalMeta, ok := indexOriginal.meta["2cf24dba5fb"]
	require.True(ok)

	indexDerivative, err := NewIndex(
		context.Background(),
		bufio.NewReader(buff),
		WithFiles([]FileMeta{&mockFile{
			relativePath: string(originalMeta.RelativePath),
			dir:          originalMeta.IsDir,
			path:         string(originalMeta.AbsolutePath),
			modTime:      originalMeta.ModTime.Add(time.Hour * 48),
			name:         originalMeta.Name,
			size:         100,
		}}),
	)
	require.NoError(err)
	assert.Len(t, indexDerivative.meta, len(indexOriginal.meta))
}
