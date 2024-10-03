package models

import (
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"third/merrors"
)

type Gallery struct {
	ID     int
	UserID int
	Title  string
}

type Image struct {
	GalleryID int
	Path      string
	Filename  string
}
type GalleryService struct {
	DB        *sql.DB
	ImagesDir string
}

func (service *GalleryService) Create(userID int, title string) (*Gallery, error) {
	gallery := Gallery{
		UserID: userID,
		Title:  title,
	}
	row := service.DB.QueryRow(`
		INSERT INTO galleries(user_id, title)
		VALUES($1, $2) 
		RETURNING id;`, gallery.UserID, gallery.Title)
	err := row.Scan(&gallery.ID)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	return &gallery, nil
}

func (service *GalleryService) ByID(id int) (*Gallery, error) {
	gallery := Gallery{
		ID: id,
	}
	row := service.DB.QueryRow(`
		SELECT user_id, title
		FROM galleries
		WHERE id = $1`, gallery.ID)
	err := row.Scan(&gallery.UserID, &gallery.Title)
	if err != nil {
		if merrors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("byID gallery: %w", err)
	}
	return &gallery, nil
}

func (service *GalleryService) ByUserID(userID int) ([]Gallery, error) {
	rows, err := service.DB.Query(`
		SELECT id, title
		FROM galleries
		WHERE user_id = $1;`, userID)
	if err != nil {
		return nil, fmt.Errorf("ByUserId gallery %w", err)
	}

	var galleries []Gallery
	for rows.Next() {
		gallery := Gallery{
			UserID: userID,
		}
		err = rows.Scan(&gallery.ID, &gallery.Title)
		if err != nil {
			return nil, fmt.Errorf("ByUserId gallery %w", err)
		}
		galleries = append(galleries, gallery)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("ByUserId gallery %w", rows.Err())
	}
	return galleries, nil
}

func (service *GalleryService) Update(gallery *Gallery) error {
	_, err := service.DB.Exec(
		`UPDATE galleries
		SET title = $2
		WHERE id = $1`, gallery.ID, gallery.Title)
	if err != nil {
		return fmt.Errorf("update Gallery %w", err)
	}
	return nil
}

func (service *GalleryService) Delete(id int) error {
	_, err := service.DB.Exec(
		`DELETE FROM galleries
		WHERE id = $1;`, id,
	)
	if err != nil {
		return fmt.Errorf("delete gallery: %w", err)
	}
	galleryDir := service.galleryDir(id)
	err = os.RemoveAll(galleryDir)
	if err != nil {
		return fmt.Errorf("delete gallery images: %w", err)
	}
	return nil
}

func (service *GalleryService) Images(galleryID int) ([]Image, error) {
	globPattern := filepath.Join(service.galleryDir(galleryID), "*")
	allFiles, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, fmt.Errorf("retrieving gallery images: %w", err)
	}
	var images []Image
	for _, file := range allFiles {
		if hasExtension(file, service.extensions()) {
			images = append(images, Image{
				Path:      file,
				GalleryID: galleryID,
				Filename:  filepath.Base(file),
			})
		}
	}
	return images, nil
}

func (service *GalleryService) CreateImage(galleryId int, filename string, contents io.ReadSeeker) error {
	err := checkContentType(contents, service.imageContentType())
	if err != nil {
		return fmt.Errorf("creating image: %w", err)
	}
	err = checkExtension(filename, service.extensions())
	if err != nil {
		return fmt.Errorf("creating image: %w", err)
	}

	galleryDir := service.galleryDir(galleryId)
	err = os.MkdirAll(galleryDir, 0755)
	if err != nil {
		return fmt.Errorf("creating gallery-%d directory %w", galleryId, err)
	}
	imagePath := filepath.Join(galleryDir, filename)
	dst, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("creating image file %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, contents)
	if err != nil {
		return fmt.Errorf("copying contents to image file %w", err)
	}
	return nil
}

func (service *GalleryService) DeleteImage(galleryId int, filename string) error {
	image, err := service.Image(galleryId, filename)
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	err = os.Remove(image.Path)
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	return nil
}
func (service *GalleryService) Image(galleryId int, filename string) (*Image, error) {
	imagePath := filepath.Join(service.galleryDir(galleryId), filename)
	_, err := os.Stat(imagePath)
	if err != nil {
		if merrors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("quering for image: %w", err)
	}
	return &Image{
		Filename:  filename,
		GalleryID: galleryId,
		Path:      imagePath,
	}, nil
}

func (service *GalleryService) extensions() []string {
	return []string{".png", ".jpg", ".jpeg", ".gif"}
}

func (service *GalleryService) imageContentType() []string {
	return []string{"image/png", "image/jpg", "image/jpeg", "image/gif"}
}

func (service *GalleryService) galleryDir(id int) string {
	imagesDir := service.ImagesDir
	if imagesDir == "" {
		imagesDir = "images"
	}
	return filepath.Join(imagesDir, fmt.Sprintf("gallery-%d", id))
}

func hasExtension(file string, extension []string) bool {
	for _, ext := range extension {
		file = strings.ToLower(file)
		ext = strings.ToLower(ext)
		if filepath.Ext(file) == ext {
			return true
		}
	}
	return false
}
