package main

import (
	"log"
	"net/http"
	"os"

	"backend/packages"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	jwtMiddleware, _ := backend.Middleware()

	r := mux.NewRouter()
	r.Handle("/list", jwtMiddleware.Handler(backend.GetList)).Methods("GET")
	r.Handle("/list/add", jwtMiddleware.Handler(backend.AddTask)).Methods("POST")
	r.Handle("/list/delete/{id}", jwtMiddleware.Handler(backend.DeleteTask)).Methods("DELETE")
	r.Handle("/list/edit/{id}", jwtMiddleware.Handler(backend.EditTask)).Methods("PUT")
	r.Handle("/list/done/{id}", jwtMiddleware.Handler(backend.DoneTask)).Methods("PUT")

	// for handling CORS
	c := cors.New(cors.Options{
		// Only add 1 value to allowed origins. Only the first one works. "*" is no exception.
		AllowedOrigins:   []string{"https://mighty-fjord-07080.herokuapp.com"},
		AllowedMethods:   []string{"GET", "DELETE", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Origin", "Accept", "Authorization"},
		AllowCredentials: true,
	})

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8000"
	}

	handler := c.Handler(r)
	log.Println("Listening on port " + port + "...")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
