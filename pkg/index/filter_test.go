package index

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIncludesFilter(t *testing.T) {
	t.Parallel()

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
	require.NoError(t, err)
	require.Len(t, passFiles, 1)
	_, ok := passFiles[paths[3].RelativePath()]
	require.True(t, ok)
}

func TestGetExcludedFiles(t *testing.T) {
	t.Parallel()

	paths := []FileMeta{
		mockFile{relativePath: "test/abc", dir: true},
		mockFile{relativePath: "test/abc/aloha.jpg"},
		mockFile{relativePath: "test/video", dir: true},
		mockFile{relativePath: "test/video/hello.mp4"},
		mockFile{relativePath: "test/video/hello1.mp4"},
		mockFile{relativePath: "test/video/good.mp4"},
		mockFile{relativePath: "test/video/abc", dir: true},
		mockFile{relativePath: "test/video/abc/salute.png"},
	}
	includePattern := "good[.]mp4$"
	excludePattern := "\\.(mp4)$"
	excludedFiles, err := getExcludedFiles(paths, includePattern, excludePattern)
	require.NoError(t, err)
	require.Len(t, excludedFiles, 2)
	_, ok := excludedFiles[paths[3].RelativePath()]
	require.True(t, ok)
	_, ok = excludedFiles[paths[4].RelativePath()]
	require.True(t, ok)
}
