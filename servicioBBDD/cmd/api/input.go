package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var rowsRejected, rowsInserted int

type tupla struct {
	Fecha         string
	Ciudad        string
	TempMax       float64
	TempMin       float64
	Precipitacion float64
	Nubosidad     float64
}

func leerCSV(path string) ([]tupla, error) {
	rowsRejected = 0
	rowsInserted = 0
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.TrimLeadingSpace = true

	var tuplas []tupla
	line := 0
	hasData := false

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error al leer CSV: %v", err)
		}

		if line == 0 {
			line++
			continue
		}

		fecha := strings.TrimSpace(record[0])
		ciudad := strings.TrimSpace(record[1])

		tmax, err1 := strconv.ParseFloat(strings.ReplaceAll(record[2], ",", "."), 64)
		tmin, err2 := strconv.ParseFloat(strings.ReplaceAll(record[3], ",", "."), 64)
		prec, err3 := strconv.ParseFloat(strings.ReplaceAll(record[4], ",", "."), 64)
		nub, err4 := strconv.ParseFloat(strings.ReplaceAll(record[5], ",", "."), 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			rowsRejected++
			continue
		}

		tuplas = append(tuplas, tupla{
			Fecha:         fecha,
			Ciudad:        ciudad,
			TempMax:       tmax,
			TempMin:       tmin,
			Precipitacion: prec,
			Nubosidad:     nub,
		})
		rowsInserted++
		hasData = true
	}

	if !hasData {
		return nil, fmt.Errorf("archivo CSV vacío o sin datos válidos después del encabezado")
	}

	return tuplas, nil
}
