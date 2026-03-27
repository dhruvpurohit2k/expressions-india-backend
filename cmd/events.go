package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) GetUpcomingEvent(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id,title,start_date,end_date FROM event WHERE status='upcoming' ORDER BY start_date LIMIT 10;`
	events := []EventListItem{}
	if err := s.db.Select(&events, query); err != nil {
		fmt.Println(err)
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

	insertMedia := `INSERT INTO media (media_type,url,s3_key) VALUES ($1,$2,$3) ON CONFLICT (url) DO UPDATE SET url=EXCLUDED.url RETURNING id`
	insertEventMedia := `INSERT INTO event_media (event_id,media_id) VALUES ($1,$2)`
	insertEvent := `INSERT INTO EVENT 
	(title,description,start_date,end_date,start_time,end_time,location,is_paid,price,perks)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`

	var perks []string
	if err := json.Unmarshal([]byte(r.FormValue("perks")), &perks); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	perksJson, err := json.Marshal(perks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var eventId string
	err = tx.QueryRow(
		insertEvent,
		r.FormValue("title"),
		r.FormValue("description"),
		r.FormValue("startDate"),
		r.FormValue("endDate"),
		r.FormValue("startTime"),
		r.FormValue("endTime"),
		r.FormValue("location"),
		r.FormValue("isPaid"),
		r.FormValue("price"),
		perksJson,
	).Scan(&eventId)
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
	fmt.Println(id)
	query := `
	SELECT 
		id,
		title,
		description,
		perks,
		start_date,
		end_date,
		start_time,
		end_time,
		location,
		is_paid,
		price
	FROM event 
	WHERE id=$1;
	`

	// temp struct to hold raw perks
	type eventRow struct {
		ID          string  `db:"id"`
		Title       string  `db:"title"`
		Description string  `db:"description"`
		Perks       []byte  `db:"perks"`
		StartDate   string  `db:"start_date"`
		EndDate     *string `db:"end_date"`
		StartTime   *string `db:"start_time"`
		EndTime     *string `db:"end_time"`
		Location    string  `db:"location"`
		IsPaid      bool    `db:"is_paid"`
		Price       *int    `db:"price"`
	}

	var row eventRow

	if err := s.db.Get(&row, query, id); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var perks []string
	if len(row.Perks) > 0 {
		if err := json.Unmarshal(row.Perks, &perks); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	event := EventDTO{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Perks:       perks,
		StartDate:   row.StartDate,
		EndDate:     row.EndDate,
		StartTime:   row.StartTime,
		EndTime:     row.EndTime,
		Location:    row.Location,
		IsPaid:      row.IsPaid,
		Price:       row.Price,
	}

	// media query (same as yours)
	linkQuery := `
	SELECT m.id, m.url
	FROM media m
	JOIN event_media em ON m.id = em.media_id
	WHERE em.event_id=$1;
	`

	var link []UploadedMedia
	if err := s.db.Select(&link, linkQuery, id); err == nil {
		event.UploadedMedia = link
	}

	// encode response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(event)
}
func (s *Server) PutEvent(w http.ResponseWriter, r *http.Request) {
	tx, err := s.db.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	id := r.PathValue("id")
	updateQuery := `UPDATE event SET title=$1, description=$2, start_date=$3, end_date=NULLIF($4,'')::DATE, start_time=NULLIF($5,'')::TIME, end_time=NULLIF($6,'')::TIME, location=$7, is_paid=$8, price=NULLIF($9,'')::INT, perks=$10 WHERE id=$11;`
	insertEventMedia := `INSERT INTO event_media (event_id,media_id) VALUES ($1,$2)`
	insertMedia := `INSERT INTO media (media_type,url,s3_key) VALUES ($1,$2,$3) ON CONFLICT (url) DO UPDATE SET url=EXCLUDED.url RETURNING id`

	var perks []string
	if err := json.Unmarshal([]byte(r.FormValue("perks")), &perks); err != nil {
		fmt.Println("Error parsing perks:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	perksJson, err := json.Marshal(perks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(updateQuery,
		r.FormValue("title"),
		r.FormValue("description"),
		r.FormValue("startDate"),
		r.FormValue("endDate"),
		r.FormValue("startTime"),
		r.FormValue("endTime"),
		r.FormValue("location"),
		r.FormValue("isPaid"),
		r.FormValue("price"),
		perksJson,
		id,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	files := r.MultipartForm.File["medias"]
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
	deleteMedia := `DELETE FROM media WHERE id=$1 RETURNING s3_key CASCADE`
	deletedIds := r.MultipartForm.Value["deletedMediaIds"]
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
