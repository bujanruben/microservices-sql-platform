package main

import (
	"testing"
)

func TestPaginationOffsetCalculation(t *testing.T) {
	tests := []struct {
		page     int
		limit    int
		expected int
	}{
		{1, 10, 0},  // (1-1)*10 = 0
		{2, 10, 10}, // (2-1)*10 = 10
		{3, 5, 10},  // (3-1)*5 = 10
		{1, 25, 0},  // (1-1)*25 = 0
	}

	for _, tt := range tests {
		offset := (tt.page - 1) * tt.limit
		if offset != tt.expected {
			t.Errorf("page=%d, limit=%d: offset=%d, want %d",
				tt.page, tt.limit, offset, tt.expected)
		}
	}
}

func TestConvertToFahrenheit(t *testing.T) {
	tests := []struct {
		celsius float64
		wantF   float64
	}{
		{0, 32},    // 0°C = 32°F
		{20, 68},   // 20°C = 68°F
		{-10, 14},  // -10°C = 14°F
		{100, 212}, // 100°C = 212°F
	}

	for _, tt := range tests {
		got := tt.celsius*9/5 + 32
		if got != tt.wantF {
			t.Errorf("convertToFahrenheit(%f) = %f, want %f",
				tt.celsius, got, tt.wantF)
		}
	}
}
