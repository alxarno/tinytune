package index

import (
	"context"
	"os"
	"testing"

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
	sample, ok := index.meta["005f6b0265"]
	require.True(ok)
	require.True(sample.IsVideo())
	require.Equal("sample_960x400_ocean_with_audio.flv", sample.Name)
	require.Equal(960, sample.Resolution.Width)
	require.Equal(400, sample.Resolution.Height)
	require.EqualValues(37038, sample.Preview.Length)
	require.EqualValues(87368, sample.Preview.Offset)
	require.EqualValues(7222114, sample.OriginSize)
	require.LessOrEqual(sample.Preview.Offset+sample.Preview.Length, uint32(len(index.data)))
}
