package preview

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCodecs(t *testing.T) {
	t.Parallel()

	sampleOutput := `(codec av1)
(codec h264)
(codec hevc)
(codec mjpeg)
(codec mpeg1video)
(codec mpeg2video)
(codec mpeg4)
(codec vc1)
(codec vp8)
(codec vp9)
`
	buff := bytes.NewBufferString(sampleOutput)
	result := formatCodecs(buff)

	//nolint:lll
	assert.Equal(t, []string{"av1", "h264", "hevc", "mjpeg", "mpeg1video", "mpeg2video", "mpeg4", "vc1", "vp8", "vp9"}, result)
}
