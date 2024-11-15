package index

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrPathMatch = errors.New("failed to match path")
var ErrIncludes = errors.New("failed filter includes files")
var ErrExcludes = errors.New("failed filter exclude files")

func (ib *indexBuilder) getExcludedFiles() (map[RelativePath]struct{}, error) {
	excluded, err := filter(ib.params.files, filterHandler(ib.params.excludePatterns))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExcludes, err)
	}

	included, err := filter(ib.params.files, filterHandler(ib.params.includePatterns))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIncludes, err)
	}

	for include := range included {
		_, ok := excluded[include]
		if ok {
			delete(excluded, include)
		}
	}

	return excluded, nil
}

func filter(items []FileMeta, filterFunc func(FileMeta) (bool, error)) (map[RelativePath]struct{}, error) {
	dst := map[RelativePath]struct{}{}

	for _, file := range items {
		if file.IsDir() {
			continue
		}

		pass, err := filterFunc(file)
		if err != nil {
			return nil, err
		}

		if pass {
			dst[RelativePath(file.RelativePath())] = struct{}{}
		}
	}

	return dst, nil
}

func filterHandler(pattern string) func(FileMeta) (bool, error) {
	patterns := strings.Split(pattern, ",")

	return func(file FileMeta) (bool, error) {
		if len(pattern) == 0 {
			return true, nil
		}

		for _, p := range patterns {
			matched, err := regexp.MatchString(p, file.RelativePath())
			if err != nil {
				return false, fmt.Errorf("%w [%v] :%w", ErrPathMatch, p, err)
			}

			if matched {
				return true, nil
			}
		}

		return false, nil
	}
}
