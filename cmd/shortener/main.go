package main

import "net/http"
import "github.com/faust8888/shortener/internal/app/handlers"

func main() {
	handler := handlers.CreateInMemoryHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handler.CreateShortURL(w, r)
		case "GET":
			handler.GetFullURL(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
