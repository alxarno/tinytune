package internal

import (
	"slices"
	"sort"
	"strconv"

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
		return a[i].OriginSize > a[j].OriginSize
	})

	return a
}

func metaSortAlphabet(slice []*index.Meta) []*index.Meta {
	sort.Slice(slice, func(first, second int) bool {
		if firstNumber, err := strconv.Atoi(slice[first].Name); err == nil {
			if secondNumber, err := strconv.Atoi(slice[second].Name); err == nil {
				return firstNumber < secondNumber
			}

			return true
		}

		return slice[first].Name < slice[second].Name
	})

	return slice
}

func metaSortAlphabetReverse(a []*index.Meta) []*index.Meta {
	result := metaSortAlphabet(a)
	slices.Reverse(result)

	return result
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

func getSorts() map[string]metaSortFunc {
	return map[string]metaSortFunc{
		"A-Z":            metaSortAlphabet,
		"Z-A":            metaSortAlphabetReverse,
		"Last Modified":  metaSortLastModified,
		"First Modified": metaSortFirstModified,
		"Type":           metaSortType,
		"Size":           metaSortSize,
	}
}
