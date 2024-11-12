package internal

import (
	"sort"

	"github.com/alxarno/tinytune/pkg/index"
)

type metaSortFunc func([]*index.Meta) []*index.Meta

func metaSortType(slice []*index.Meta) []*index.Meta {
	sort.Slice(slice, func(i, j int) bool {
		if slice[i].Type == slice[j].Type {
			return slice[i].Name > slice[j].Name
		}

		return slice[i].Type > slice[j].Type
	})

	return slice
}

func metaSortSize(a []*index.Meta) []*index.Meta {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Preview.Length > a[j].Preview.Length
	})

	return a
}

func metaSortAlphabet(a []*index.Meta) []*index.Meta {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Name < a[j].Name
	})

	return a
}

func metaSortAlphabetReverse(a []*index.Meta) []*index.Meta {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Name > a[j].Name
	})

	return a
}

func metaSortLastModified(a []*index.Meta) []*index.Meta {
	sort.Slice(a, func(i, j int) bool {
		return a[i].ModTime.Unix() > a[j].ModTime.Unix()
	})

	return a
}

func metaSortFirstModified(a []*index.Meta) []*index.Meta {
	sort.Slice(a, func(i, j int) bool {
		return a[i].ModTime.Unix() < a[j].ModTime.Unix()
	})

	return a
}
