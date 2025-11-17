package main

import (
	"testing"
	"time"
)

func TestGenerateCacheKey_ConsistencyAndUniqueness(t *testing.T) {
	// Misma entrada → misma clave
	key1 := GenerateCacheKey("Madrid", "2024-01-15", 5, "C", "daily")
	key2 := GenerateCacheKey("Madrid", "2024-01-15", 5, "C", "daily")

	if key1 != key2 {
		t.Error("Claves idénticas deberían generar el mismo hash")
	}

	// Diferente entrada → diferente clave
	key3 := GenerateCacheKey("Barcelona", "2024-01-15", 5, "C", "daily")
	if key1 == key3 {
		t.Error("Claves diferentes deberían generar hashes diferentes")
	}

	// Formato correcto (SHA1 = 40 chars hex)
	if len(key1) != 40 {
		t.Errorf("Hash SHA1 debería tener 40 caracteres, tiene %d", len(key1))
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache()

	cache.Set("key1", "value1", 100*time.Millisecond)

	// Inmediatamente debería existir
	if _, found := cache.Get("key1"); !found {
		t.Error("Item debería existir inmediatamente después de Set")
	}

	// Después de expirar debería desaparecer
	time.Sleep(150 * time.Millisecond)
	if _, found := cache.Get("key1"); found {
		t.Error("Item debería expirar después del TTL")
	}
}
