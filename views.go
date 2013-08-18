package main

import (
	ae "appengine"
	ds "appengine/datastore"
	"net/http"
)

func getPost(r *http.Request, slug string) (ae.Context, *ds.Key, *Post, error) {
	c := ae.NewContext(r)
	k := ds.NewKey(c, "post", slug, 0, nil)
	p := NewPost(slug)
	err := ds.Get(c, k, p)
	return c, k, p, err
}

func viewPost(w http.ResponseWriter, r *http.Request, slug string) {
	_, _, p, err := getPost(r, slug)
	if err != nil {
		serveError(w, err)
		return
	}
	render(w, "post", p)
}

func editPost(w http.ResponseWriter, r *http.Request, slug string) {
	_, _, p, err := getPost(r, slug)
	if err != nil && err != ds.ErrNoSuchEntity {
		serveError(w, err)
		return
	}
	render(w, "edit-post", p)
}

func savePost(w http.ResponseWriter, r *http.Request, slug string) {
	c, k, p, err := getPost(r, slug)
	create := err == ds.ErrNoSuchEntity
	if err != nil && !create {
		serveError(w, err)
		return
	}
	p.init(r, create)
	_, err = ds.Put(c, k, p)
	if err != nil {
		serveError(w, err)
		return
	}
	http.Redirect(w, r, "/post/"+slug, http.StatusFound)
}

func viewHome(w http.ResponseWriter, r *http.Request) {
	c := ae.NewContext(r)
	q := ds.NewQuery("post").Order("-Created").Limit(10)
	var ps []*Post
	_, err := q.GetAll(c, &ps)
	if err != nil {
		serveError(w, err)
	}
	render(w, "home", ps)
}

func viewAbout(w http.ResponseWriter, r *http.Request) {
	render(w, "about", nil)
}

func init() {
	http.HandleFunc(handler("/post/", viewPost))
	http.HandleFunc(handler("/edit-post/", editPost))
	http.HandleFunc(handler("/save-post/", savePost))
	http.HandleFunc("/about/", viewAbout)
	http.HandleFunc("/", viewHome)
}
