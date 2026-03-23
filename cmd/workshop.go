package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) GetWorkshopType(w http.ResponseWriter, r *http.Request) {
	query := `SELECT name FROM workshop_type;`
	var workshopTypes []string
	err := s.db.Select(&workshopTypes, query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"types": workshopTypes})
	// w.WriteHeader(http.StatusOK)

}

func (s *Server) GetWorkshops(w http.ResponseWriter, r *http.Request) {
	query := `SELECT w.id, w.title, w.description, wt.name AS workshop_type, w.start_date, w.end_date FROM workshop w JOIN workshop_type wt ON wt.id = w.workshop_type ORDER BY wt.id ASC, w.start_date ;`

	workshops := []WorkshopDTO{}

	err := s.db.Select(&workshops, query)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workshops)
}

func (s *Server) PostWorkshop(w http.ResponseWriter, r *http.Request) {
	tx, err := s.db.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	insertWorkshop := `INSERT INTO workshop (title,description,workshop_type,start_date,end_date,registration_link) VALUES ($1,$2,$3,$4,NULLIF($5,'')::DATE,NULLIF($6,'')) RETURNING id`
	insertMedia := `INSERT INTO media (media_type,url,s3_key) VALUES ($1,$2,$3) ON CONFLICT (url) DO UPDATE SET url=EXCLUDED.url RETURNING id`
	insertWorkshopMedia := `INSERT INTO workshop_media (workshop_id,media_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
	var workshopId string
	err = tx.QueryRow(insertWorkshop,
		r.FormValue("title"),
		r.FormValue("description"),
		r.FormValue("workshopType"),
		r.FormValue("startDate"),
		r.FormValue("endDate"),
		r.FormValue("registrationLink"),
	).Scan(&workshopId)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	files := r.MultipartForm.File["medias"]
	for _, fileHeader := range files {
		f, _ := fileHeader.Open()
		defer f.Close()

		location, s3Key, err := s.uploadTos3IO(f, fileHeader.Filename, "events")
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var mediaId string
		err = tx.QueryRow(insertMedia, "s3", location, s3Key).Scan(&mediaId)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(insertWorkshopMedia, workshopId, mediaId)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "success"})

}

func (s *Server) GetWorkshop(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	query := `SELECT id,title,description,workshop_type,start_date,end_date,registration_link FROM workshop WHERE id=$1;`
	workshop := WorkshopDTO{}
	if err := s.db.Get(&workshop, query, id); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	linkQuery := `
			SELECT m.id,m.url
			FROM media m
			JOIN workshop_media wm ON m.id = wm.media_id
		WHERE wm.workshop_id=$1;
		`
	var link []UploadedMedia
	s.db.Select(&link, linkQuery, id)
	workshop.UploadedMedia = link
	data, _ := json.Marshal(workshop)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) PutWorkshop(w http.ResponseWriter, r *http.Request) {
	tx, err := s.db.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	id := r.PathValue("id")
	updateQuery := `UPDATE workshop SET title=$1, description=$2, registration_link=NULLIF($3,''), workshop_type=$4, start_date=$5, end_date=NULLIF($6,'')::DATE WHERE id=$7;`
	insertWorkshopMedia := `INSERT INTO workshop_media (workshop_id,media_id) VALUES ($1,$2)`
	insertMedia := `INSERT INTO media (media_type,url,s3_key) VALUES ($1,$2,$3) ON CONFLICT (url) DO UPDATE SET url=EXCLUDED.url RETURNING id`

	// endDateStr := r.FormValue("endDate")
	// var endDate any
	// if endDateStr == "" {
	// 	endDate = nil
	// } else {
	// 	endDate = endDateStr
	// }

	_, err = tx.Exec(updateQuery, r.FormValue("title"), r.FormValue("description"), r.FormValue("registrationLink"), r.FormValue("workshopType"), r.FormValue("startDate"), r.FormValue("registrationLink"), id)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	files := r.MultipartForm.File["medias"]
	for _, fileHeader := range files {
		f, _ := fileHeader.Open()
		defer f.Close()

		location, s3key, err := s.uploadTos3IO(f, fileHeader.Filename, "workshop")
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var mediaId string
		err = tx.QueryRow(insertMedia, "s3", location, s3key).Scan(&mediaId)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(insertWorkshopMedia, id, mediaId)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
	deleteMedia := `DELETE FROM media WHERE id=$1 RETURNING s3_key`
	deletedIds := r.MultipartForm.Value["deletedMediaIds"]
	for _, id := range deletedIds {
		var s3Key string
		err := tx.QueryRow(deleteMedia, id).Scan(&s3Key)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = s.DeleteFromS3(s3Key)
		if err != nil {
			fmt.Println("Couldn't delete", s3Key)
		}
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})

}
