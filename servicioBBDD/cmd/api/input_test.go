package main

import (
	"os"
	"testing"
)

func TestLeerCSV_ValidData(t *testing.T) {
	csvContent := `fecha;ciudad;temp_max;temp_min;precipitacion;nubosidad
2024-01-15;Madrid;15,5;5,2;0,0;45,8
2024-01-16;Madrid;16,2;6,1;2,1;67,3`

	tmpFile, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(csvContent)
	tmpFile.Close()

	tuplas, err := leerCSV(tmpFile.Name())
	if err != nil {
		t.Fatalf("leerCSV falló: %v", err)
	}

	if len(tuplas) != 2 {
		t.Errorf("Esperaba 2 tuplas, obtuve %d", len(tuplas))
	}
	if rowsInserted != 2 {
		t.Errorf("Esperaba 2 filas insertadas, obtuve %d", rowsInserted)
	}
	if rowsRejected != 0 {
		t.Errorf("Esperaba 0 filas rechazadas, obtuve %d", rowsRejected)
	}
}

func TestLeerCSV_InvalidData(t *testing.T) {
	csvContent := `fecha;ciudad;temp_max;temp_min;precipitacion;nubosidad
2024-01-15;Madrid;INVALIDO;5,2;0,0;45,8
2024-01-16;Madrid;16,2;6,1;2,1;67,3`

	tmpFile, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(csvContent)
	tmpFile.Close()

	tuplas, err := leerCSV(tmpFile.Name())
	if err != nil {
		t.Fatalf("leerCSV falló: %v", err)
	}

	if len(tuplas) != 1 {
		t.Errorf("Esperaba 1 tupla válida, obtuve %d", len(tuplas))
	}
	if rowsRejected != 1 {
		t.Errorf("Esperaba 1 fila rechazada, obtuve %d", rowsRejected)
	}
}

func TestLeerCSV_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()

	_, err = leerCSV(tmpFile.Name())
	if err == nil {
		t.Error("Esperaba error con archivo vacío")
	}
}
