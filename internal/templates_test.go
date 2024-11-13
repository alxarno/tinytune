package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTemplatesFuncs(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "01:15:24", durationPrint(time.Hour*1+time.Minute*15+time.Second*24))
	assert.Equal(t, "15:24", durationPrint(time.Minute*15+time.Second*24))
	assert.Equal(t, "00:24", durationPrint(time.Second*24))
}
