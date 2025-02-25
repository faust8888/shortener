package main

import (
	"github.com/faust8888/shortener/internal/app/handlers"
	"net/http"
)

func main() {
	router := handlers.CreateRouter(handlers.CreateInMemoryHandler())

	if err := http.ListenAndServe(`:8080`, router); err != nil {
		panic(err)
	}
}
