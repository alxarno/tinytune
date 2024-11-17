package internal

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustCompile(t *testing.T, list string) []*regexp.Regexp {
	t.Helper()

	patterns := strings.Split(list, ",")
	compiled := make([]*regexp.Regexp, len(patterns))

	for i, pattern := range patterns {
		reg, err := regexp.Compile(pattern)
		require.NoError(t, err)

		compiled[i] = reg
	}

	return compiled
}

func TestIncludesFilter(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	paths, err := NewCrawlerOS("../test/").Scan("index.tinytune")
	require.NoError(err)

	pattern := "\\.(mp4)$"
	passFiles := filter(paths, filterHandler(mustCompile(t, pattern)))
	require.Len(passFiles, 3)
	_, ok := passFiles["../test/sample.mp4"]
	require.True(ok)

	_, ok = passFiles["../test/video/sample.mp4"]
	require.True(ok)
}

func TestGetExcludedFiles(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	files, err := NewCrawlerOS("../test/").Scan("index.tinytune")
	require.NoError(err)

	includePatterns := mustCompile(t, "video/sample[.]mp4$")
	excludePatterns := mustCompile(t, "\\.(mp4)$")

	excludedFiles := GetExcludedFiles(files, includePatterns, excludePatterns)
	require.Len(excludedFiles, 2)
	_, ok := excludedFiles["../test/sample.mp4"]
	require.True(ok)
}
