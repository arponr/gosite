package main

import (
	"net/http"
	"time"
)

type Post struct {
	Slug, Title, ImgUrl string
	Preview, Content    []byte
	Created, Edited     time.Time
	Public              bool
}

func (p *Post) init(r *http.Request, create bool) {
	p.Public = r.FormValue("public") == "true"
	p.Title = r.FormValue("title")
	p.ImgUrl = r.FormValue("imgurl")
	p.Preview = []byte(r.FormValue("preview"))
	p.Content = []byte(r.FormValue("content"))
	if create {
		p.Created = time.Now()
	}
	p.Edited = time.Now()
}

type List struct {
	P    []*Post
	Path string
	Page int
	More bool
}
