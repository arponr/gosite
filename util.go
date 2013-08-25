package main

import (
	"github.com/russross/blackfriday"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
	fmtDate     = "2 Jan 2006"
	fmtDatetime = "2 Jan 2006, 3:04pm"
)

func date(t time.Time) interface{}     { return t.Format(fmtDate) }
func datetime(t time.Time) interface{} { return t.Format(fmtDatetime) }
func safe(s string) interface{}        { return template.HTML(s) }
func add(i, j int) interface{}         { return i + j }

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
	files = append(files, "html/base.html")
	return template.Must(template.New("").Funcs(template.FuncMap{
		"add":      add,
		"date":     date,
		"datetime": datetime,
		"markdown": markdown,
		"safe":     safe,
	}).ParseFiles(files...))
}

var templates = map[string]*template.Template{
	"dne":      buildTemplate("html/dne.html"),
	"error":    buildTemplate("html/error.html"),
	"editpost": buildTemplate("html/editpost.html"),
	"post":     buildTemplate("html/post.html"),
	"list":     buildTemplate("html/list.html"),
	"about":    buildTemplate("html/about.html"),
}

func render(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates[tmpl].ExecuteTemplate(w, "base.html", data)
	if err != nil {
		serveError(w, err)
	}
}

func serveDne(w http.ResponseWriter)              { render(w, "dne", nil) }
func serveError(w http.ResponseWriter, err error) { render(w, "error", nil) }

var slugger = regexp.MustCompile(`^[-a-zA-Z0-9]+$`)

type slugView func(http.ResponseWriter, *http.Request, string)

func slugHandler(path string, v slugView) (string, http.HandlerFunc) {
	return path, func(w http.ResponseWriter, r *http.Request) {
		slug := r.URL.Path[len(path):]
		if !slugger.MatchString(slug) {
			serveDne(w)
			return
		}
		v(w, r, slug)
	}
}

type pageView func(http.ResponseWriter, *http.Request, string, bool, int)

func pageHandler(path string, v pageView) (string, http.HandlerFunc) {
	return path, func(w http.ResponseWriter, r *http.Request) {
		home := len(r.URL.Path) == len(path)
		var page int
		if home {
			page = 1
		} else {
			i, err := strconv.Atoi(r.URL.Path[len(path):])
			if err != nil || i <= 0 {
				serveDne(w)
				return
			}
			page = i
		}
		v(w, r, path, home, page)
	}
}

func staticHandler(path string, v http.HandlerFunc) (string, http.HandlerFunc) {
	return path, func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) != len(path) {
			serveDne(w)
			return
		}
		v(w, r)
	}
}
