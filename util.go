package main

import (
	"html/template"
	"net/http"
	"regexp"
	"time"

	"github.com/russross/blackfriday"
)

const (
	fmtDate     = "2 Jan 2006"
	fmtDatetime = "2 Jan 2006, 3:04pm"
)

func date(t time.Time) interface{}     { return t.Format(fmtDate) }
func datetime(t time.Time) interface{} { return t.Format(fmtDatetime) }
func safe(s string) interface{}        { return template.HTML(s) }

var (
	censor   = regexp.MustCompile(`\$\$[^\$]+\$\$|\$[^\$]+\$`)
	uncensor = regexp.MustCompile(`\$+`)
)

func replace(vals [][]byte) func([]byte) []byte {
	i := -1
	return func(b []byte) []byte {
		i++
		return vals[i]
	}
}

func markdown(input []byte) interface{} {
	matches := censor.FindAll(input, -1)
	tex := make([][]byte, len(matches))
	for i, m := range matches {
		tex[i] = make([]byte, len(m))
		for j := range m {
			tex[i][j], m[j] = m[j], '$'
		}
	}
	output := blackfriday.MarkdownCommon(input)
	return uncensor.ReplaceAllFunc(output, replace(tex))
}

func buildTemplate(files ...string) *template.Template {
	files = append(files, "html/dne.html", "html/base.html")
	return template.Must(template.New("").Funcs(template.FuncMap{
		"date":     date,
		"datetime": datetime,
		"markdown": markdown,
		"safe":     safe,
	}).ParseFiles(files...))
}

var (
	validator = regexp.MustCompile(`^[-a-zA-Z0-9]+$`)
	templates = map[string]*template.Template{
		"edit-post": buildTemplate("html/edit-post.html"),
		"post":      buildTemplate("html/post.html"),
		"home":      buildTemplate("html/home.html"),
		"about":     buildTemplate("html/about.html"),
	}
)

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func render(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates[tmpl].ExecuteTemplate(w, "base.html", data)
	if err != nil {
		serveError(w, err)
	}
}

type view func(http.ResponseWriter, *http.Request, string)

func handler(path string, v view) (string, http.HandlerFunc) {
	return path, func(w http.ResponseWriter, r *http.Request) {
		slug := r.URL.Path[len(path):]
		if !validator.MatchString(slug) {
			render(w, "dne", nil)
			return
		}
		v(w, r, slug)
	}
}
