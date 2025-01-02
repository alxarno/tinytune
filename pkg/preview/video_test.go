package preview

import (
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
			DataLength: 33492,
			Duration:   time.Second * 5,
			Width:      1280,
			Height:     720,
			Hash:       "b5039ea0fe138cd39b7c5b1b1bad83324fcb02f73862094a3408bc74ea522b36",
		},
		{
			Name:       ".flv",
			SourcePath: "../../test/video/sample_960x400_ocean_with_audio.flv",
			DataLength: 12448,
			Duration:   time.Second * 46,
			Width:      960,
			Height:     400,
			Hash:       "e65ce8f7043b05951ec56cfd1f141398745b66478247978f6075e695c8631af7",
		},
		{
			Name:       "short",
			SourcePath: "../../test/short.mp4",
			DataLength: 33762,
			Duration:   time.Second * 3,
			Width:      1280,
			Height:     720,
			Hash:       "05c84f197db0752646e50fedc1156e165a2a226ef5c207aa72c2543202c9dad3",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			require := require.New(t)
			preview, err := videoPreview(testCase.SourcePath, VideoParams{timeout: time.Minute})
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
