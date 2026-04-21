package main

import (
	"encoding/json"
	"net/http"
)

type API struct {
	store *Store
}

func NewAPI(store *Store) *API {
	return &API{store: store}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /set/{key}", a.handleSet)
	mux.HandleFunc("GET /get/{key}", a.handleGet)
	mux.HandleFunc("DELETE /delete/{key}", a.handleDelete)
}

func (a *API) handleSet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	var body struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := a.store.Set(key, body.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (a *API) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	value, err := a.store.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"value": value})
}

func (a *API) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if err := a.store.Delete(key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
