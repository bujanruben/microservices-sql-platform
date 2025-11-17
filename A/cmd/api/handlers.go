package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func handleIngestCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error leyendo archivo: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	tmpDir := os.TempDir()
	tmpPath := fmt.Sprintf("%s/%s", tmpDir, header.Filename)

	out, err := os.Create(tmpPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creando archivo temporal: %v", err), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, fmt.Sprintf("Error copiando archivo: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Archivo CSV guardado temporalmente en: %s", tmpPath)

	f, err := os.Open(tmpPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error abriendo archivo temporal: %v", err), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		http.Error(w, fmt.Sprintf("Error calculando checksum: %v", err), http.StatusInternalServerError)
		return
	}
	fileChecksum := fmt.Sprintf("sha256:%x", h.Sum(nil))

	tuplas, err := leerCSV(tmpPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error procesando CSV: %v", err), http.StatusInternalServerError)
		return
	}

	db, err := connectDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error conectando a DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		log.Fatal(err)
	}

	err = insertTuplas(db, tuplas)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error insertando en DB: %v", err), http.StatusInternalServerError)
		return
	}

	elapsedMs := time.Since(start).Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"rows_inserted": %d, "rows_rejected": %d, "elapsed_ms": %d, "file_checksum": "%s"}`,
		rowsInserted, rowsRejected, elapsedMs, fileChecksum)
}

func handleGetCities(w http.ResponseWriter, r *http.Request) {
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Error conectando a la base de datos", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	ciudades, err := getCiudades(db)
	if err != nil {
		http.Error(w, "Error consultando ciudades", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ciudades)
}

func handleGetRaw(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if city == "" || from == "" || to == "" {
		http.Error(w, "Parámetros 'city', 'from' y 'to' son obligatorios", http.StatusBadRequest)
		return
	}

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Error conectando a la base de datos", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	registros, err := getTuplasPaginadas(db, city, from, to, page, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error consultando datos: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(registros)
}
