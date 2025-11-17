package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func connectDB() (*sql.DB, error) {
	connStr := "host=meteo-db port=5432 user=postgres password=postgres dbname=meteo sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error conectando a la base de datos: %v", err)
	}

	return db, nil
}

func initDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS weather_data (
		id SERIAL PRIMARY KEY,
		fecha DATE NOT NULL,
		ciudad TEXT NOT NULL,
		temp_max DOUBLE PRECISION,
		temp_min DOUBLE PRECISION,
		precipitacion DOUBLE PRECISION,
		nubosidad DOUBLE PRECISION
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creando tabla: %v", err)
	}
	return nil
}

func insertTuplas(db *sql.DB, tuplas []tupla) error {
	stmt, err := db.Prepare(`
		INSERT INTO weather_data (fecha, ciudad, temp_max, temp_min, precipitacion, nubosidad)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tuplas {
		_, err := stmt.Exec(t.Fecha, t.Ciudad, t.TempMax, t.TempMin, t.Precipitacion, t.Nubosidad)
		if err != nil {
			return err
		}
	}
	return nil
}

func getCiudades(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT ciudad FROM weather_data ORDER BY ciudad;`)
	if err != nil {
		return nil, fmt.Errorf("error en la query: %v", err)
	}
	defer rows.Close()
	var ciudades []string

	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, fmt.Errorf("error escaneando fila: %v", err)
		}
		ciudades = append(ciudades, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error tras recorrer filas: %v", err)
	}

	return ciudades, nil
}

func getTuplasPaginadas(db *sql.DB, city, from, to string, page, limit int) ([]tupla, error) {
	offset := (page - 1) * limit

	query := `
        SELECT fecha, ciudad, temp_max, temp_min, precipitacion, nubosidad
        FROM weather_data
        WHERE ciudad = $1
          AND fecha BETWEEN $2 AND $3
        ORDER BY fecha
        LIMIT $4 OFFSET $5
    `
	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, fmt.Errorf("formato de fecha 'from' inválido: %v", err)
	}
	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil, fmt.Errorf("formato de fecha 'to' inválido: %v", err)
	}

	rows, err := db.Query(query, city, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error en consulta SQL: %v", err)
	}
	defer rows.Close()

	var resultados []tupla
	for rows.Next() {
		var t tupla
		err := rows.Scan(&t.Fecha, &t.Ciudad, &t.TempMax, &t.TempMin, &t.Precipitacion, &t.Nubosidad)
		if err != nil {
			return nil, fmt.Errorf("error leyendo fila: %v", err)
		}
		resultados = append(resultados, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando resultados: %v", err)
	}

	return resultados, nil
}
