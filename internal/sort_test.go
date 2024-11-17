package internal

import (
	"testing"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSorts(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		input  []*index.Meta
		output []*index.Meta
	}{
		{
			name: "A-Z",
			input: []*index.Meta{
				{Name: "cba"},
				{Name: "bca"},
				{Name: "abc"},
			},
			output: []*index.Meta{
				{Name: "abc"},
				{Name: "bca"},
				{Name: "cba"},
			},
		},
		{
			name: "Z-A",
			input: []*index.Meta{
				{Name: "abc"},
				{Name: "bca"},
				{Name: "cba"},
			},
			output: []*index.Meta{
				{Name: "cba"},
				{Name: "bca"},
				{Name: "abc"},
			},
		},
		{
			name: "Last Modified",
			input: []*index.Meta{
				{ModTime: time.Date(2021, time.May, 12, 21, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2021, time.May, 12, 22, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2021, time.May, 12, 23, 00, 00, 00, time.UTC)},
			},
			output: []*index.Meta{
				{ModTime: time.Date(2021, time.May, 12, 23, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2021, time.May, 12, 22, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2021, time.May, 12, 21, 00, 00, 00, time.UTC)},
			},
		},
		{
			name: "First Modified",
			input: []*index.Meta{
				{ModTime: time.Date(2021, time.May, 12, 23, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2021, time.May, 12, 22, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2021, time.May, 12, 21, 00, 00, 00, time.UTC)},
			},
			output: []*index.Meta{
				{ModTime: time.Date(2021, time.May, 12, 21, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2021, time.May, 12, 22, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2021, time.May, 12, 23, 00, 00, 00, time.UTC)},
			},
		},
		{
			name: "Type",
			input: []*index.Meta{
				{Type: index.ContentTypeImage},
				{Type: index.ContentTypeDir},
				{Type: index.ContentTypeOther},
				{Type: index.ContentTypeVideo},
			},
			output: []*index.Meta{
				{Type: index.ContentTypeDir},
				{Type: index.ContentTypeImage},
				{Type: index.ContentTypeVideo},
				{Type: index.ContentTypeOther},
			},
		},
		{
			name: "Size",
			input: []*index.Meta{
				{OriginSize: 1024 * 2},
				{OriginSize: 1024},
				{OriginSize: 1024 * 10},
				{OriginSize: 1024 * 3},
			},
			output: []*index.Meta{
				{OriginSize: 1024 * 10},
				{OriginSize: 1024 * 3},
				{OriginSize: 1024 * 2},
				{OriginSize: 1024},
			},
		},
	}

	sorts := getSorts()

	for _, tCase := range testCases {
		t.Run(tCase.name, func(test *testing.T) {
			test.Parallel()

			require := require.New(test)
			sort, ok := sorts[tCase.name]
			require.True(ok)
			sort(tCase.input)
			assert.ElementsMatch(test, tCase.input, tCase.output)
		})
	}
}
