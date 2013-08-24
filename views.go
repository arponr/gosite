package main

import (
	"appengine/user"
	"net/http"

	ae "appengine"
	ds "appengine/datastore"
)

func getPost(r *http.Request, slug string) (ae.Context, *ds.Key, *Post, error) {
	c := ae.NewContext(r)
	k := ds.NewKey(c, "post", slug, 0, nil)
	p := NewPost(slug)
	err := ds.Get(c, k, p)
	return c, k, p, err
}

func viewPost(w http.ResponseWriter, r *http.Request, slug string) {
	c, _, p, err := getPost(r, slug)
	if err == ds.ErrNoSuchEntity {
		serveDne(w)
		return
	}
	if err != nil {
		serveError(w, err)
		return
	}
	if !p.Public && !user.IsAdmin(c) {
		serveDne(w)
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
	render(w, "editpost", p)
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

func viewAbout(w http.ResponseWriter, r *http.Request) {
	render(w, "about", nil)
}

func viewBlog(w http.ResponseWriter, r *http.Request) {
	c := ae.NewContext(r)
	q := ds.NewQuery("post").Order("-Created").Limit(10)
	var ps []*Post
	_, err := q.GetAll(c, &ps)
	if err != nil {
		serveError(w, err)
	}
	render(w, "blog", ps)
}

func viewHome(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) != 1 {
		serveDne(w)
		return
	}
	viewBlog(w, r)
}

func init() {
	http.HandleFunc(handler("/post/", viewPost))
	http.HandleFunc("/about/", viewAbout)
	http.HandleFunc("/", viewHome)
	http.HandleFunc(handler("/admin/editpost/", editPost))
	http.HandleFunc(handler("/admin/savepost/", savePost))
}
