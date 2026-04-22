package main

import (
	"log"
	"net/http"

	"store/engines/lsm"
)

func main() {
	engine, err := lsm.NewLSMEngine("data/lsm")
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

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
