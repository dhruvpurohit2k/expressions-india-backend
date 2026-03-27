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

	query := `
    SELECT w.id, w.title, wt.name AS workshop_type, w.start_date, w.end_date 
    FROM workshop w 
    JOIN workshop_type wt ON wt.id = w.workshop_type 
    WHERE w.start_date >= CURRENT_DATE + INTERVAL '1 day'
    ORDER BY wt.id ASC, w.start_date
`

	workshops := []WorkshopListDTO{}

	err := s.db.Select(&workshops, query)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	listOfWorkshops := make(map[string][]WorkshopListDTO)

	for _, workshop := range workshops {
		_, ok := listOfWorkshops[workshop.WorkshopType]
		if !ok {
			listOfWorkshops[workshop.WorkshopType] = make([]WorkshopListDTO, 0)
		}
		listOfWorkshops[workshop.WorkshopType] = append(listOfWorkshops[workshop.WorkshopType], workshop)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]map[string][]WorkshopListDTO{"data": listOfWorkshops})
}

func (s *Server) PostWorkshop(w http.ResponseWriter, r *http.Request) {
	tx, err := s.db.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	insertWorkshop := `INSERT INTO workshop (title,description,workshop_type,start_date,end_date,start_time,end_time,location,is_paid,price,perks) VALUES ($1,$2,$3,$4,NULLIF($5,'')::DATE,NULLIF($6,'')::TIME,NULLIF($7,'')::TIME,$8,$9,NULLIF($10,'')::INT,$11) RETURNING id`
	insertMedia := `INSERT INTO media (media_type,url,s3_key) VALUES ($1,$2,$3) ON CONFLICT (url) DO UPDATE SET url=EXCLUDED.url RETURNING id`
	insertWorkshopMedia := `INSERT INTO workshop_media (workshop_id,media_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`

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

	var workshopId string
	err = tx.QueryRow(insertWorkshop,
		r.FormValue("title"),
		r.FormValue("description"),
		r.FormValue("workshopType"),
		r.FormValue("startDate"),
		r.FormValue("endDate"),
		r.FormValue("startTime"),
		r.FormValue("endTime"),
		r.FormValue("location"),
		r.FormValue("isPaid"),
		r.FormValue("price"),
		perksJson,
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
		price,
		workshop_type
	FROM workshop
	WHERE id=$1;
	`

	// temp struct to hold raw perks
	type workshopRow struct {
		ID           string  `db:"id"`
		Title        string  `db:"title"`
		Description  string  `db:"description"`
		Perks        []byte  `db:"perks"`
		StartDate    string  `db:"start_date"`
		EndDate      *string `db:"end_date"`
		StartTime    *string `db:"start_time"`
		EndTime      *string `db:"end_time"`
		Location     string  `db:"location"`
		IsPaid       bool    `db:"is_paid"`
		Price        *int    `db:"price"`
		WorkshopType int     `db:"workshop_type"`
	}

	var row workshopRow

	if err := s.db.Get(&row, query, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var perks []string
	if len(row.Perks) > 0 {
		if err := json.Unmarshal(row.Perks, &perks); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	event := WorkshopDTO{
		ID:           row.ID,
		Title:        row.Title,
		Description:  row.Description,
		Perks:        perks,
		StartDate:    row.StartDate,
		EndDate:      row.EndDate,
		StartTime:    row.StartTime,
		EndTime:      row.EndTime,
		Location:     row.Location,
		IsPaid:       row.IsPaid,
		Price:        row.Price,
		WorkshopType: row.WorkshopType,
	}

	// media query (same as yours)
	linkQuery := `
	SELECT m.id, m.url
	FROM media m
	JOIN workshop_media wm ON m.id = wm.media_id
	WHERE wm.workshop_id=$1;
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

func (s *Server) PutWorkshop(w http.ResponseWriter, r *http.Request) {
	tx, err := s.db.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	id := r.PathValue("id")
	updateQuery := `UPDATE workshop SET title=$1, description=$2, workshop_type=$3, start_date=$4, end_date=NULLIF($5,'')::DATE, start_time=NULLIF($6,'')::TIME, end_time=NULLIF($7,'')::TIME, location=$8, is_paid=$9, price=NULLIF($10,'')::INT, perks=$11 WHERE id=$12;`
	insertWorkshopMedia := `INSERT INTO workshop_media (workshop_id,media_id) VALUES ($1,$2)`
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
		r.FormValue("workshopType"), 
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
