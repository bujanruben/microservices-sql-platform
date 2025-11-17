package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Tupla struct {
	Fecha         string
	Ciudad        string
	TempMax       float64
	TempMin       float64
	Precipitacion float64
	Nubosidad     float64
}

func ConnectDB() (*sql.DB, error) {
	connStr := "host=meteo-db port=5432 user=postgres password=postgres dbname=meteo sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error conectando a PostgreSQL: %v", err)
	}

	return db, nil
}
func GetWeatherData(db *sql.DB, city string, date string, days int, unit, agg string) ([]Tupla, error) {
	start, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("formato de fecha inválido: %v", err)
	}
	end := start.AddDate(0, 0, days-1)
	endDate := end.Format("2006-01-02")

	var query string
	var args []interface{}

	switch agg {
	case "daily":
		query = `
			SELECT 
				fecha,
				ciudad,
				MIN(temp_min) as temp_min,
				MAX(temp_max) as temp_max,
				(AVG(temp_max) + AVG(temp_min)) / 2 as temp_avg,
				SUM(precipitacion) as precipitacion,
				AVG(nubosidad) as nubosidad
			FROM weather_data
			WHERE ciudad = $1 AND fecha BETWEEN $2 AND $3
			GROUP BY fecha, ciudad
			ORDER BY fecha
		`
		args = []interface{}{city, date, endDate}

	case "rolling7":
		query = `
			SELECT 
				fecha,
				ciudad,
				AVG((temp_max + temp_min) / 2) OVER (
					ORDER BY fecha ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
				) as temp_avg_7d,
				AVG(nubosidad) OVER (
					ORDER BY fecha ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
				) as nubosidad_7d,
				SUM(precipitacion) OVER (
					ORDER BY fecha ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
				) as precipitacion_7d,
				0 as temp_min,
				0 as temp_max
			FROM weather_data
			WHERE ciudad = $1 AND fecha BETWEEN $2 AND $3
			ORDER BY fecha
		`
		args = []interface{}{city, date, endDate}

	default:
		query = `
			SELECT fecha, ciudad, temp_max, temp_min, precipitacion, nubosidad
			FROM weather_data
			WHERE ciudad = $1 AND fecha BETWEEN $2 AND $3
			ORDER BY fecha
		`
		args = []interface{}{city, date, endDate}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error en consulta SQL: %v", err)
	}
	defer rows.Close()

	var resultados []Tupla

	for rows.Next() {
		var t Tupla

		switch agg {
		case "rolling7":

			var tempAvg7d, nubosidad7d, precipitacion7d float64
			err := rows.Scan(&t.Fecha, &t.Ciudad, &tempAvg7d, &nubosidad7d, &precipitacion7d, &t.TempMin, &t.TempMax)
			if err != nil {
				return nil, fmt.Errorf("error leyendo fila (rolling7): %v", err)
			}

			t.TempMax = tempAvg7d
			t.TempMin = tempAvg7d
			t.Nubosidad = nubosidad7d
			t.Precipitacion = precipitacion7d

		case "daily":

			var tempAvg float64
			err := rows.Scan(&t.Fecha, &t.Ciudad, &t.TempMin, &t.TempMax, &tempAvg, &t.Precipitacion, &t.Nubosidad)
			if err != nil {
				return nil, fmt.Errorf("error leyendo fila (daily): %v", err)
			}

		default:
			err := rows.Scan(&t.Fecha, &t.Ciudad, &t.TempMax, &t.TempMin, &t.Precipitacion, &t.Nubosidad)
			if err != nil {
				return nil, fmt.Errorf("error leyendo fila: %v", err)
			}
		}

		resultados = append(resultados, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando resultados: %v", err)
	}

	if unit == "F" {
		resultados = convertToFahrenheit(resultados)
	}

	return resultados, nil
}

func convertToFahrenheit(data []Tupla) []Tupla {
	converted := make([]Tupla, len(data))
	for i, d := range data {
		converted[i] = Tupla{
			Fecha:         d.Fecha,
			Ciudad:        d.Ciudad,
			TempMax:       d.TempMax*9/5 + 32,
			TempMin:       d.TempMin*9/5 + 32,
			Precipitacion: d.Precipitacion,
			Nubosidad:     d.Nubosidad,
		}
	}
	return converted
}
