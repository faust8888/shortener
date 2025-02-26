package main

import (
	"flag"
	"github.com/faust8888/shortener/cmd/config"
	"github.com/faust8888/shortener/internal/app/handlers"
	"net/http"
)

func main() {
	flag.Parse()
	router := handlers.CreateRouter(handlers.CreateInMemoryHandler())

	if err := http.ListenAndServe(config.Config.ServerAddress, router); err != nil {
		panic(err)
	}
}
