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
func markdown(b []byte) interface{}    { return blackfriday.MarkdownCommon(b) }

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
	validator = regexp.MustCompile("^[-a-zA-Z0-9]+$")
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
