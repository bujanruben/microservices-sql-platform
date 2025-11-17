package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWeatherHandler_ParameterValidation(t *testing.T) {
	cache = NewCache()
	tests := []struct {
		name           string
		path           string
		query          string
		expectedStatus int
	}{
		{
			name:           "missing_city_and_date",
			path:           "/weather/",
			query:          "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing_date",
			path:           "/weather/Madrid",
			query:          "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_days_too_high",
			path:           "/weather/Madrid",
			query:          "date=2024-01-15&days=11",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_days_too_low",
			path:           "/weather/Madrid",
			query:          "date=2024-01-15&days=0",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_unit",
			path:           "/weather/Madrid",
			query:          "date=2024-01-15&unit=X",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_agg",
			path:           "/weather/Madrid",
			query:          "date=2024-01-15&agg=invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid_params",
			path:           "/weather/Madrid",
			query:          "date=2024-01-15&days=5&unit=C&agg=daily",
			expectedStatus: http.StatusOK, // O 503 si no hay BD
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.path
			if tt.query != "" {
				url += "?" + tt.query
			}

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			WeatherHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Para %s: status = %d, quiere %d", tt.name, rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestExponentialRetry_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	err := exponentialRetry(func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("error temporal")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Debería tener éxito después de 3 intentos, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Debería hacer 3 intentos, got: %d", attempts)
	}
}

func TestCalculateDelay_ExponentialGrowth(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},
		{1, 3 * time.Second},  // 1 * 3^1
		{2, 9 * time.Second},  // 1 * 3^2
		{3, 27 * time.Second}, // 1 * 3^3
		{4, 27 * time.Second}, // Capped at maxRetryDelay
	}

	for _, tt := range tests {
		got := calculateDelay(tt.attempt)
		if got != tt.want {
			t.Errorf("calculateDelay(%d) = %v, quiere %v", tt.attempt, got, tt.want)
		}
	}
}
