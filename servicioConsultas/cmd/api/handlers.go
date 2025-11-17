package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	maxRetries     = 4
	baseRetryDelay = 1 * time.Second
	maxRetryDelay  = 27 * time.Second
)

func exponentialRetry(operation func() error) error {
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		if attempt < maxRetries-1 {
			delay := calculateDelay(attempt)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("operación falló después de %d intentos: %v", maxRetries, err)
}

func calculateDelay(attempt int) time.Duration {
	delay := baseRetryDelay * time.Duration(math.Pow(3, float64(attempt)))
	if delay > maxRetryDelay {
		return maxRetryDelay
	}
	return delay
}

func WeatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	city := strings.TrimPrefix(path, "/weather/")
	date := r.URL.Query().Get("date")
	daysStr := r.URL.Query().Get("days")
	unit := r.URL.Query().Get("unit")
	agg := r.URL.Query().Get("agg")

	if city == "" || date == "" {
		http.Error(w, `{"error": "Ciudad y fecha son obligatorios"}`, http.StatusBadRequest)
		return
	}

	days := 5
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	if days < 1 || days > 10 {
		http.Error(w, `{"error": "Days debe estar entre 1 y 10"}`, http.StatusBadRequest)
		return
	}

	if unit == "" {
		unit = "C"
	}
	if unit != "C" && unit != "F" {
		http.Error(w, `{"error": "Unit debe ser C o F"}`, http.StatusBadRequest)
		return
	}

	if agg != "" && agg != "daily" && agg != "rolling7" {
		http.Error(w, `{"error": "Agg debe ser daily o rolling7"}`, http.StatusBadRequest)
		return
	}

	cacheKey := GenerateCacheKey(city, date, days, unit, agg)

	if cached, found := cache.Get(cacheKey); found {
		sendWeatherResponse(w, city, date, days, unit, agg, cached.([]Tupla))
		return
	}

	var db *sql.DB
	var err error

	err = exponentialRetry(func() error {
		db, err = ConnectDB()
		return err
	})

	if err != nil {
		http.Error(w, `{"error": "Servicio de base de datos no disponible y no hay datos en caché"}`, http.StatusServiceUnavailable)
		return
	}
	defer db.Close()

	var weatherData []Tupla
	err = exponentialRetry(func() error {
		weatherData, err = GetWeatherData(db, city, date, days, unit, agg)
		return err
	})

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "No se pudieron obtener datos: %v"}`, err), http.StatusInternalServerError)
		return
	}

	cache.Set(cacheKey, weatherData, 10*time.Minute)

	sendWeatherResponse(w, city, date, days, unit, agg, weatherData)
}

func sendWeatherResponse(w http.ResponseWriter, city, date string, days int, unit, agg string, weatherData []Tupla) {
	start, _ := time.Parse("2006-01-02", date)
	end := start.AddDate(0, 0, days-1)
	endDate := end.Format("2006-01-02")

	response := map[string]interface{}{
		"city":  city,
		"from":  date,
		"to":    endDate,
		"days":  days,
		"unit":  unit,
		"agg":   agg,
		"data":  weatherData,
		"count": len(weatherData),
	}

	if len(weatherData) == 0 {
		response["message"] = "No se encontraron datos"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
