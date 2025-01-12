package internal

import (
	"errors"
	"regexp"
	"strings"

	"github.com/alxarno/tinytune/pkg/index"
)

var (
	ErrPathMatch = errors.New("failed to match path")
	ErrIncludes  = errors.New("failed filter includes files")
	ErrExcludes  = errors.New("failed filter exclude files")
)

func GetExcludedFiles(files []index.FileMeta, included, excluded []*regexp.Regexp) map[string]struct{} {
	excludedFiles := filter(files, filterHandler(excluded))
	includedFiles := filter(files, filterHandler(included))

	for include := range includedFiles {
		delete(excludedFiles, include)
	}

	return excludedFiles
}

func GetIncludedFiles(files []index.FileMeta, included []*regexp.Regexp) map[string]struct{} {
	return filter(files, filterHandler(included))
}

func filter(items []index.FileMeta, filterFunc func(index.FileMeta) bool) map[string]struct{} {
	dst := map[string]struct{}{}

	for _, file := range items {
		if file.IsDir() {
			continue
		}

		if filterFunc(file) {
			dst[file.Path()] = struct{}{}
		}
	}

	return dst
}

func filterHandler(patterns []*regexp.Regexp) func(index.FileMeta) bool {
	return func(file index.FileMeta) bool {
		for _, p := range patterns {
			if p.MatchString(strings.ToLower(file.Path())) {
				return true
			}
		}

		return false
	}
}
