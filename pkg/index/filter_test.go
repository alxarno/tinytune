package index

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIncludesFilter(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	paths := []FileMeta{
		mockFile{relativePath: "test/abc", dir: true},
		mockFile{relativePath: "test/abc/aloha.jpg"},
		mockFile{relativePath: "test/video", dir: true},
		mockFile{relativePath: "test/video/hello.mp4"},
		mockFile{relativePath: "test/video/abc", dir: true},
		mockFile{relativePath: "test/video/abc/salute.png"},
	}
	pattern := "\\.(mp4)$"
	passFiles, err := filter(paths, filterHandler(pattern))
	require.NoError(err)
	require.Len(passFiles, 1)
	_, ok := passFiles[RelativePath(paths[3].RelativePath())]
	require.True(ok)
}

func TestGetExcludedFiles(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	files := []FileMeta{
		mockFile{relativePath: "test/abc", dir: true},
		mockFile{relativePath: "test/abc/aloha.jpg"},
		mockFile{relativePath: "test/video", dir: true},
		mockFile{relativePath: "test/video/hello.mp4"},
		mockFile{relativePath: "test/video/hello1.mp4"},
		mockFile{relativePath: "test/video/good.mp4"},
		mockFile{relativePath: "test/video/abc", dir: true},
		mockFile{relativePath: "test/video/abc/salute.png"},
	}
	indexBuilder := indexBuilder{params: indexBuilderParams{
		includePatterns: "good[.]mp4$",
		excludePatterns: "\\.(mp4)$",
		files:           files,
	}}
	excludedFiles, err := indexBuilder.getExcludedFiles()
	require.NoError(err)
	require.Len(excludedFiles, 2)
	_, ok := excludedFiles[RelativePath(files[3].RelativePath())]
	require.True(ok)
	_, ok = excludedFiles[RelativePath(files[4].RelativePath())]
	require.True(ok)
}
