package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrawlerOS(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	files, err := NewCrawlerOS("../test").Scan("../test/index.tinytune")
	require.NoError(err)
	require.Len(files, 15)
}
