package bytesutil

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriterCount(t *testing.T) {
	t.Parallel()

	buff := new(bytes.Buffer)
	writer := NewWriterCounter(buff)
	_, err := writer.Write([]byte("abcd"))
	require.NoError(t, err)
	require.EqualValues(t, 4, writer.Count())
}
