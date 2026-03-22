package main

import (
	"fmt"
	"net/http"
)

func main() {
	middlerwares := setupMiddleWares(enableCORS)
	server := createServer()
	if server == nil {
		fmt.Println("Failed to open server")
		return
	}
	fmt.Println("Listening on port 5000")
	http.ListenAndServe(":5000", middlerwares(server))
}
