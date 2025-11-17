package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/ingest/csv", handleIngestCSV)
	http.HandleFunc("/cities", handleGetCities)
	http.HandleFunc("/raw", handleGetRaw)

	fmt.Println("Servidor iniciado en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
