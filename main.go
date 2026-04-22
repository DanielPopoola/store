package main

import (
	"log"
	"net/http"

	"store/engines"
)

func main() {
	engine, err := engines.NewFileEngine("data.log")
	if err != nil {
		log.Fatal(err)
	}

	store := NewStore(engine)

	api := NewAPI(store)
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	addr := ":8080"
	log.Printf("Server running on %s\n", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
