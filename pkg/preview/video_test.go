package preview

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	ErrPreviewLengthMismatch = errors.New("preview length mismatch")
)

func TestPreviewVideo(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name       string
		SourcePath string
		DataLength int
		Duration   time.Duration
		Width      int
		Height     int
		Hash       string
	}{
		{
			Name:       ".mp4",
			SourcePath: "../../test/sample.mp4",
			DataLength: 35134,
			Duration:   time.Second * 5,
			Width:      1280,
			Height:     720,
			Hash:       "7eda8710470a1547e43e115a7dc98d9a24bef80d23cd4014f0848e17bb0af1b9",
		},
		{
			Name:       ".flv",
			SourcePath: "../../test/video/sample_960x400_ocean_with_audio.flv",
			DataLength: 12476,
			Duration:   time.Second * 46,
			Width:      960,
			Height:     400,
			Hash:       "daecc12d47bc6b36b0f8ff828552b9da6da98bea77727422a9d2a7cae8ef96cf",
		},
		{
			Name:       "short",
			SourcePath: "../../test/short.mp4",
			DataLength: 35228,
			Duration:   time.Second * 3,
			Width:      1280,
			Height:     720,
			Hash:       "8a7a907329c6670540c89a78bb6722d362c9dd635894cd3f0e1b1202891d5402",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			require := require.New(t)
			preview, err := videoPreview(context.Background(), testCase.SourcePath, VideoParams{timeout: time.Minute})
			require.NoError(err)
			require.Len(
				preview.Data(),
				testCase.DataLength,
				fmt.Errorf("%w: %d != %d", ErrPreviewLengthMismatch, len(preview.Data()), testCase.DataLength),
			)
			require.EqualValues(testCase.Duration, preview.Duration())
			width, height := preview.Resolution()
			require.Equal(testCase.Width, width)
			require.Equal(testCase.Height, height)

			hash := sha256.Sum256(preview.Data())
			require.Equal(testCase.Hash, hex.EncodeToString(hash[:]))
		})
	}
}
