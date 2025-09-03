package handlers

import (
	"encoding/json"
	"net/http"
)

type Progression struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

var progressions = make(map[string]Progression)

func CreateProgression(w http.ResponseWriter, r *http.Request) {
	var progression Progression
	if err := json.NewDecoder(r.Body).Decode(&progression); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	progressions[progression.ID] = progression
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(progression)
}

func GetProgression(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	progression, exists := progressions[id]
	if !exists {
		http.Error(w, "Progression not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(progression)
}

func ListProgressions(w http.ResponseWriter, r *http.Request) {
	var progressionList []Progression
	for _, progression := range progressions {
		progressionList = append(progressionList, progression)
	}
	json.NewEncoder(w).Encode(progressionList)
}