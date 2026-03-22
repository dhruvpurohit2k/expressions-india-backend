package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
	insertWorkshopMedia := `INSERT INTO workshop_media (workshop_id,media_id) VALUES ($1,$2)`
	var workshopId string
	// workshopType, err := strconv.Atoi(r.FormValue("workshopType"))
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
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
	return

}
