package logic

import (
	"net/http"
	"time"
)

type Post struct {
	Slug, Title      string
	Preview, Content []byte
	Created, Edited  time.Time
}

func NewPost(slug string) *Post {
	return &Post{Slug: slug}
}

func (p *Post) init(r *http.Request, create bool) {
	p.Title = r.FormValue("title")
	p.Preview = []byte(r.FormValue("preview"))
	p.Content = []byte(r.FormValue("content"))
	if create {
		p.Created = time.Now()
	}
	p.Edited = time.Now()
}
