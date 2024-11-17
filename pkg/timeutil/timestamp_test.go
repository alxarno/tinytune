package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimestampString(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	duration := time.Hour*23 + time.Minute*7 + time.Second*9
	require.Equal("23:07:09", String(duration))

	duration = time.Minute*7 + time.Second*9
	require.Equal("07:09", String(duration))

	duration = time.Second * 10
	require.Equal("00:10", String(duration))
}
