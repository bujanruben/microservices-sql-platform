package main

import (
	"log"
	"net/http"
)

var cache *Cache

func main() {

	cache = NewCache()

	http.HandleFunc("/weather/", WeatherHandler)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
