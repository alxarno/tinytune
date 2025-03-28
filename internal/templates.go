package internal

import (
	"html/template"
	"io/fs"
	"path"
	"strings"

	"github.com/alxarno/tinytune/pkg/timeutil"
)

func loadTemplates(src fs.FS, streaming map[string]struct{}) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	funcs := template.FuncMap{
		"ext": extension,
		"eqMinusOne": func(x int, y int) bool {
			return x == y-1
		},
		"dur":       timeutil.String,
		"streaming": getStreaming(streaming),
	}
	templates["index.html"] = template.Must(template.New("index.html").Funcs(funcs).ParseFS(src, "*.html"))

	return templates
}

func extension(name string) string {
	extension := path.Ext(name)
	if extension != "" {
		return strings.ToLower(extension[1:])
	}

	return ""
}

func getStreaming(files map[string]struct{}) func(path string) bool {
	return func(path string) bool {
		_, ok := files[path]

		return ok
	}
}
