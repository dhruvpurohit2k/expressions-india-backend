package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) GetEventList(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id,title,start_date,end_date FROM event LIMIT 10;`
	events := []EventListItem{}
	if err := s.db.Select(&events, query); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (s *Server) PostEvent(w http.ResponseWriter, r *http.Request) {
	tx, err := s.db.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	endDateStr := r.FormValue("endDate")

	var endDate any

	if endDateStr == "" {
		endDate = nil
	} else {
		endDate = endDateStr
	}

	insertMedia := `INSERT INTO media (media_type,url,s3_key) VALUES ($1,$2,$3) ON CONFLICT (url) DO UPDATE SET url=EXCLUDED.url RETURNING id`
	insertEventMedia := `INSERT INTO event_media (event_id,media_id) VALUES ($1,$2)`
	insertEvent := `INSERT INTO EVENT (title,description,start_date,end_date) VALUES ($1,$2,$3,$4) RETURNING id`
	var eventId string
	err = tx.QueryRow(insertEvent, r.FormValue("title"), r.FormValue("description"), r.FormValue("startDate"), endDate).Scan(&eventId)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	files := r.MultipartForm.File["images"]
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

		_, err = tx.Exec(insertEventMedia, eventId, mediaId)
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

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	query := `SELECT id,title,description,start_date,end_date FROM event WHERE id=$1;`
	event := EventDTO{}
	if err := s.db.Get(&event, query, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	linkQuery := `
			SELECT m.id,m.url
			FROM media m
			JOIN event_media em ON m.id = em.media_id
		WHERE em.event_id=$1;
		`
	var link []MediaLink
	s.db.Select(&link, linkQuery, id)
	event.MediaLink = link
	data, _ := json.Marshal(event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
func (s *Server) PutEvent(w http.ResponseWriter, r *http.Request) {
	tx, err := s.db.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	id := r.PathValue("id")
	updateQuery := `UPDATE event SET title=$1, description=$2, start_date=$3, end_date=$4 WHERE id=$5;`
	insertEventMedia := `INSERT INTO event_media (event_id,media_id) VALUES ($1,$2)`
	insertMedia := `INSERT INTO media (media_type,url,s3_key) VALUES ($1,$2,$3) ON CONFLICT (url) DO UPDATE SET url=EXCLUDED.url RETURNING id`
	endDateStr := r.FormValue("endDate")
	var endDate any
	if endDateStr == "" {
		endDate = nil
	} else {
		endDate = endDateStr
	}

	_, err = tx.Exec(updateQuery, r.FormValue("title"), r.FormValue("description"), r.FormValue("startDate"), endDate, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	files := r.MultipartForm.File["images"]
	for _, fileHeader := range files {
		f, _ := fileHeader.Open()
		defer f.Close()

		location, s3key, err := s.uploadTos3IO(f, fileHeader.Filename, "events")
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

		_, err = tx.Exec(insertEventMedia, id, mediaId)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
	deleteMedia := `DELETE FROM media WHERE id=$1 RETURNING s3_key`
	deletedIds := r.MultipartForm.Value["deletedIds"]
	fmt.Println(deletedIds)
	for _, id := range deletedIds {
		fmt.Println("DELETING", id)
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
