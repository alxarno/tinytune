package internal

import (
	"errors"
	"regexp"

	"github.com/alxarno/tinytune/pkg/index"
)

var ErrPathMatch = errors.New("failed to match path")
var ErrIncludes = errors.New("failed filter includes files")
var ErrExcludes = errors.New("failed filter exclude files")

func GetExcludedFiles(files []index.FileMeta, included, excluded []*regexp.Regexp) map[string]struct{} {
	excludedFiles := filter(files, filterHandler(excluded))
	includedFiles := filter(files, filterHandler(included))

	for include := range includedFiles {
		delete(excludedFiles, include)
	}

	return excludedFiles
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
			if p.MatchString(file.Path()) {
				return true
			}
		}

		return false
	}
}
