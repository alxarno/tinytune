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

func getExcludedFiles(items []FileMeta, includePattern string, excludePatter string) (map[string]struct{}, error) {
	excluded, err := filter(items, filterHandler(excludePatter))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExcludes, err)
	}

	included, err := filter(items, filterHandler(includePattern))
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

func filter(items []FileMeta, filterFunc func(FileMeta) (bool, error)) (map[string]struct{}, error) {
	dst := map[string]struct{}{}

	for _, file := range items {
		if file.IsDir() {
			continue
		}

		pass, err := filterFunc(file)
		if err != nil {
			return nil, err
		}

		if pass {
			dst[file.RelativePath()] = struct{}{}
		}
	}

	return dst, nil
}

func filterHandler(pattern string) func(FileMeta) (bool, error) {
	patterns := strings.Split(pattern, ",")

	return func(fm FileMeta) (bool, error) {
		for _, p := range patterns {
			matched, err := regexp.MatchString(p, fm.RelativePath())
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

// func getExcludes(pattern string) func(FileMeta) (bool, error) {
// 	patterns := strings.Split(pattern, ",")

// 	return func(fm FileMeta) (bool, error) {
// 		for _, p := range patterns {
// 			matched, err := regexp.MatchString(p, fm.RelativePath())
// 			if err != nil {
// 				return false, fmt.Errorf("%w [%v] :%w", ErrPathMatch, p, err)
// 			}

// 			if matched {
// 				return true, nil
// 			}
// 		}

// 		return false, nil
// 	}
// }
