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
	p := &Post{Slug: slug}
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
	create := err == ds.ErrNoSuchEntity || !p.Public
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

const perPage = 10

var qs = map[string]*ds.Query{
	"/":             ds.NewQuery("post").Filter("Public =", true).Order("-Created"),
	"/admin/queue/": ds.NewQuery("post").Filter("Public =", false).Order("-Edited"),
}

func viewList(w http.ResponseWriter, r *http.Request, path string, home bool, page int) {
	q := qs[path].Offset((page - 1) * perPage).Limit(perPage + 1)
	c := ae.NewContext(r)
	list := &List{Path: path, Page: page}
	_, err := q.GetAll(c, &list.P)
	if err != nil {
		serveError(w, err)
		return
	}
	if !home && len(list.P) == 0 {
		serveDne(w)
		return
	}
	list.More = len(list.P) == perPage+1
	if list.More {
		list.P = list.P[:perPage]
	}
	render(w, "list", list)
}

func init() {
	http.HandleFunc(pageHandler("/admin/queue/", viewList))
	http.HandleFunc(slugHandler("/admin/editpost/", editPost))
	http.HandleFunc(slugHandler("/admin/savepost/", savePost))

	http.HandleFunc(staticHandler("/about/", viewAbout))
	http.HandleFunc(slugHandler("/post/", viewPost))
	http.HandleFunc(pageHandler("/", viewList))
}
