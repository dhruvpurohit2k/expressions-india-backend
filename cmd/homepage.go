package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) GetHomePageImageClient(w http.ResponseWriter, r *http.Request) {
	var imageUrls []string
	query := `SELECT url FROM home_page_image;`
	s.db.Select(&imageUrls, query)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(imageUrls)
	return
}
func (s *Server) GetHomePageImage(w http.ResponseWriter, r *http.Request) {
	homePageImages := []struct {
		Id   string `json:"id" db:"id"`
		Url  string `json:"url" db:"url"`
		Type string `json:"type"`
	}{}
	query := `SELECT id,url FROM home_page_image;`
	s.db.Select(&homePageImages, query)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(homePageImages)
	return
}

func (s *Server) UpdateHomePageImage(w http.ResponseWriter, r *http.Request) {
	tx, err := s.db.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
	}
	defer tx.Rollback()
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	deletedIds := r.MultipartForm.Value["deletedImageIds"]
	images := r.MultipartForm.File["images"]
	fmt.Println(deletedIds)
	fmt.Println(images)
	if len(images) > 5 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	deletedImagesQuery := `DELETE FROM home_page_image WHERE id=$1 RETURNING s3_key;`
	imageQuery := `INSERT INTO home_page_image (url,s3_key) VALUES ($1,$2);`
	for _, id := range deletedIds {
		var key string
		err := tx.QueryRow(deletedImagesQuery, id).Scan(&key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = s.DeleteFromS3(key)
		if err != nil {
			fmt.Println("FAILED TO DELETE IMAGE")
		}
	}
	for _, image := range images {
		f, _ := image.Open()
		defer f.Close()
		location, s3Key, err := s.uploadTos3IO(f, image.Filename, "homepage")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec(imageQuery, location, s3Key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})

}
