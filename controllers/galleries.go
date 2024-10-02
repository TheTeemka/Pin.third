package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"third/context"
	"third/models"

	"github.com/go-chi/chi/v5"
)

type Galleries struct {
	Templates struct {
		Show  template
		New   template
		Edit  template
		Index template
	}
	GalleryService *models.GalleryService
}

func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
	type RGallery struct {
		ID    int
		Title string
	}
	var data struct {
		RGalleries []RGallery
	}
	user := context.User(r.Context())
	galleries, err := g.GalleryService.ByUserID(user.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	for _, gallery := range galleries {
		data.RGalleries = append(data.RGalleries, RGallery{
			ID:    gallery.ID,
			Title: gallery.Title,
		})
	}

	g.Templates.Index.Execute(w, r, data)
}

func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}
	data.Title = r.FormValue("title")
	g.Templates.New.Execute(w, r, data)
}

func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title  string
		UserId int
	}
	data.Title = r.FormValue("title")
	data.UserId = context.User(r.Context()).ID

	gallery, err := g.GalleryService.Create(data.UserId, data.Title)
	if err != nil {
		g.Templates.New.Execute(w, r, data, err)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r, userMustOwnFunction)
	if err != nil {
		return
	}
	data := struct {
		ID    int
		Title string
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}

	g.Templates.Edit.Execute(w, r, data)
}

func (g Galleries) ProcessEdit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r, userMustOwnFunction)
	if err != nil {
		return
	}

	gallery.Title = r.FormValue("title")
	err = g.GalleryService.Update(gallery)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	var data struct {
		ID     int
		Title  string
		Images []string
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	for i := 0; i < 20; i++ {
		w, h := rand.Intn(500)+200, rand.Intn(500)+200
		catImageUrl := fmt.Sprintf("https://picsum.photos/%d/%d", w, h)
		data.Images = append(data.Images, catImageUrl)
	}
	g.Templates.Show.Execute(w, r, data)
}

func (g Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r, userMustOwnFunction)
	if err != nil {
		return
	}

	err = g.GalleryService.Delete(gallery.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

type galleryOpt func(w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error

func (g Galleries) galleryByID(w http.ResponseWriter, r *http.Request, opts ...galleryOpt) (*models.Gallery, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return nil, err
	}

	gallery, err := g.GalleryService.ByID(id)
	if err != nil {
		if err == models.ErrNotFound {
			http.Error(w, "Gallery Not Found", http.StatusNotFound)
			return nil, err
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return nil, err
	}

	for _, opt := range opts {
		err = opt(w, r, gallery)
		if err != nil {
			return nil, err
		}
	}
	return gallery, nil
}

func userMustOwnFunction(w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error {
	user := context.User(r.Context())
	if user.ID != gallery.UserID {
		http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
		return fmt.Errorf("user does not have access to this gallery")
	}
	return nil
}
