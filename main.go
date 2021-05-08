package main

import (
	"log"
	"net/http"

	"./packages"
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
		AllowedOrigins:   []string{"http://localhost:3001", "*"},
		AllowedMethods:   []string{"GET", "DELETE", "POST", "PUT", "OPTIONS", "*"},
		AllowedHeaders:   []string{"Content-Type", "Origin", "Accept", "Authorization", "*"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)
	log.Println("Listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", handler))
}
