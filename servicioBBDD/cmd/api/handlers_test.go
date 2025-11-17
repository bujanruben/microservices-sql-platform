package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleIngestCSV_MethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("GET", "/ingest/csv", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleIngestCSV)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Status incorrecto: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestHandleIngestCSV_NoFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, err := http.NewRequest("POST", "/ingest/csv", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handleIngestCSV(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Status incorrecto: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestHandleGetRaw_MissingParameters(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"sin city", "/raw?from=2024-01-01&to=2024-01-10"},
		{"sin from", "/raw?city=Madrid&to=2024-01-10"},
		{"sin to", "/raw?city=Madrid&from=2024-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handleGetRaw(rr, req)

			if status := rr.Code; status != http.StatusBadRequest {
				t.Errorf("%s: status = %d, want %d", tt.name, rr.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleGetRaw_DefaultPaginationValues(t *testing.T) {
	req, err := http.NewRequest("GET", "/raw?city=Madrid&from=2024-01-01&to=2024-01-10", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handleGetRaw(rr, req)

	if rr.Code == http.StatusBadRequest {
		t.Errorf("Parámetros deberían ser válidos, got status %d", rr.Code)
	}
}
