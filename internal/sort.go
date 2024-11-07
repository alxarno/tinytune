package internal

import (
	"sort"

	"github.com/alxarno/tinytune/pkg/index"
)

type metaSortFunc func([]*index.IndexMeta) []*index.IndexMeta

func metaSortType(a []*index.IndexMeta) []*index.IndexMeta {
	sort.Slice(a, func(i, j int) bool {
		if a[i].Type == a[j].Type {
			return a[i].Name > a[j].Name
		}
		return a[i].Type > a[j].Type
	})
	return a
}

func metaSortSize(a []*index.IndexMeta) []*index.IndexMeta {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Preview.Length > a[j].Preview.Length
	})
	return a
}

func metaSortAlphabet(a []*index.IndexMeta) []*index.IndexMeta {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Name < a[j].Name
	})
	return a
}

func metaSortAlphabetReverse(a []*index.IndexMeta) []*index.IndexMeta {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Name > a[j].Name
	})
	return a
}

func metaSortLastModified(a []*index.IndexMeta) []*index.IndexMeta {
	sort.Slice(a, func(i, j int) bool {
		return a[i].ModTime.Unix() > a[j].ModTime.Unix()
	})
	return a
}

func metaSortFirstModified(a []*index.IndexMeta) []*index.IndexMeta {
	sort.Slice(a, func(i, j int) bool {
		return a[i].ModTime.Unix() < a[j].ModTime.Unix()
	})
	return a
}
