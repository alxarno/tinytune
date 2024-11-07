package internal

import (
	"sort"

	"github.com/alxarno/tinytune/pkg/index"
)

func metaSortType(a []*index.IndexMeta) []*index.IndexMeta {
	sort.Slice(a, func(i, j int) bool {
		if a[i].Type == a[j].Type {
			return a[i].Name > a[j].Name
		}
		return a[i].Type > a[j].Type
	})
	return a
}
