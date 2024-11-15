package internal

import (
	"testing"
	"time"

	"github.com/alxarno/tinytune/pkg/timeutil"
	"github.com/stretchr/testify/assert"
)

func TestTemplatesFuncs(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.Equal("01:15:24", timeutil.String(time.Hour*1+time.Minute*15+time.Second*24))
	assert.Equal("15:24", timeutil.String(time.Minute*15+time.Second*24))
	assert.Equal("00:24", timeutil.String(time.Second*24))
}
