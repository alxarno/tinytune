package bytesutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrettyByteSize(t *testing.T) {
	t.Parallel()

	size := 1024*30 + 7
	require.Equal(t, "30KB", PrettyByteSize(size))
	size = 1024 * 30
	require.Equal(t, "30KB", PrettyByteSize(size))
	size = 1024*1024*1024 + 1024*30 + 7
	require.Equal(t, "1GB", PrettyByteSize(size))
}
