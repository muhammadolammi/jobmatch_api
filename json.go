package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling payload to data %v", err)
		w.WriteHeader(500)
	}
	w.WriteHeader(code)
	_, err = w.Write(data)
	if err != nil {
		log.Printf("error writing data to response %v", err)
		w.WriteHeader(500)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJson(w, code, map[string]string{"error": message})

}
