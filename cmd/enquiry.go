package main

import (
	"fmt"
	"net/http"
)

func (s *Server) PostEnquiry(w http.ResponseWriter, r *http.Request) {
	tx, err := s.db.Beginx()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	insertEnquiry := `INSERT INTO enquiry (name,designation,contact_number,email_id,body) VALUES ($1,$2,$3,$4,$5);`

	_, err = tx.Exec(
		insertEnquiry,
		r.FormValue("name"),
		r.FormValue("designation"),
		r.FormValue("contactNumber"),
		r.FormValue("email"),
		r.FormValue("enquiry"),
	)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx.Commit()
	return

}
