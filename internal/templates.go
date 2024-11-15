package internal

import (
	"html/template"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/alxarno/tinytune/pkg/timeutil"
)

func loadTemplates(src fs.FS) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	funcs := template.FuncMap{
		"ext":    extension,
		"width":  width,
		"height": height,
		"eqMinusOne": func(x int, y int) bool {
			return x == y-1
		},
		"dur":       timeutil.String,
		"streaming": streaming,
	}
	templates["index.html"] = template.Must(template.New("index.html").Funcs(funcs).ParseFS(src, "*.html"))

	return templates
}

func extension(name string) string {
	extension := path.Ext(name)
	if extension != "" {
		return extension[1:]
	}

	return ""
}

func width(res string) string {
	return strings.Split(res, "x")[0]
}

func height(res string) string {
	return strings.Split(res, "x")[1]
}

func streaming(path string) bool {
	streamingFormats := []string{"avi", "f4v", "flv"}
	ext := filepath.Ext(path)
	minExtensionLength := 2

	if len(ext) < minExtensionLength {
		return false
	}

	if slices.Contains(streamingFormats, ext[1:]) {
		return true
	}

	return false
}
