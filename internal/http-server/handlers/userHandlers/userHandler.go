package userHandlers

import (
	"encoding/json"
	"net/http"
	"time"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		getHandler(w, r)
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(time.Now())
}
