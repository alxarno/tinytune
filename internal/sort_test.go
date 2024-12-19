package internal

import (
	"testing"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
	"github.com/stretchr/testify/require"
)

//nolint:dupl
func TestSorts(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		input  []*index.Meta
		output []*index.Meta
		equal  func(first, second *index.Meta) bool
	}{
		{
			name:  "A-Z",
			equal: func(first, second *index.Meta) bool { return first.Name == second.Name },
			input: []*index.Meta{
				{Name: "cba"},
				{Name: "bca"},
				{Name: "abc"},
				{Name: "1"},
				{Name: "30"},
				{Name: "200"},
				{Name: "31"},
				{Name: "4"},
			},
			output: []*index.Meta{
				{Name: "1"},
				{Name: "4"},
				{Name: "30"},
				{Name: "31"},
				{Name: "200"},
				{Name: "abc"},
				{Name: "bca"},
				{Name: "cba"},
			},
		},
		{
			name:  "Z-A",
			equal: func(first, second *index.Meta) bool { return first.Name == second.Name },
			input: []*index.Meta{
				{Name: "cba"},
				{Name: "bca"},
				{Name: "abc"},
				{Name: "1"},
				{Name: "30"},
				{Name: "200"},
				{Name: "31"},
				{Name: "4"},
			},
			output: []*index.Meta{
				{Name: "cba"},
				{Name: "bca"},
				{Name: "abc"},
				{Name: "200"},
				{Name: "31"},
				{Name: "30"},
				{Name: "4"},
				{Name: "1"},
			},
		},
		{
			name:  "Last Modified",
			equal: func(first, second *index.Meta) bool { return first.ModTime == second.ModTime },
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
			name:  "First Modified",
			equal: func(first, second *index.Meta) bool { return first.ModTime == second.ModTime },
			input: []*index.Meta{
				{ModTime: time.Date(2022, time.May, 13, 23, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2022, time.May, 13, 22, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2022, time.May, 13, 21, 00, 00, 00, time.UTC)},
			},
			output: []*index.Meta{
				{ModTime: time.Date(2022, time.May, 13, 21, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2022, time.May, 13, 22, 00, 00, 00, time.UTC)},
				{ModTime: time.Date(2022, time.May, 13, 23, 00, 00, 00, time.UTC)},
			},
		},
		{
			name:  "Type",
			equal: func(first, second *index.Meta) bool { return first.Type == second.Type },
			input: []*index.Meta{
				{Type: index.ContentTypeImage},
				{Type: index.ContentTypeDir},
				{Type: index.ContentTypeOther},
				{Type: index.ContentTypeVideo},
			},
			output: []*index.Meta{
				{Type: index.ContentTypeDir},
				{Type: index.ContentTypeOther},
				{Type: index.ContentTypeImage},
				{Type: index.ContentTypeVideo},
			},
		},
		{
			name:  "Size",
			equal: func(first, second *index.Meta) bool { return first.OriginSize == second.OriginSize },
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

			for i := range tCase.input {
				require.True(tCase.equal(tCase.input[i], tCase.output[i]))
			}
		})
	}
}
