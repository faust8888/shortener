package main

import (
	"fmt"
	"github.com/faust8888/shortener/cmd/config"
	"github.com/faust8888/shortener/internal/app/handlers"
	"net/http"
)

func main() {
	fmt.Println("### Configuring server ###")
	config.LoadConfig()
	router := handlers.CreateRouter(handlers.CreateInMemoryHandler())
	fmt.Printf("Starting server on %s\n\n", config.Cfg.ServerAddress)

	if err := http.ListenAndServe(config.Cfg.ServerAddress, router); err != nil {
		panic(err)
	}
}
