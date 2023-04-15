package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	"github.com/pokgak/mygoapp/handlers"
	"github.com/pokgak/mygoapp/utils"
)

func main() {
	var err error

	utils.InitTraceProvider()

	// initialize database
	utils.InitDatabase()

	// configure router
	r := mux.NewRouter()

	r.HandleFunc("/users", handlers.UsersPostHandler).Methods(http.MethodPost)
	r.HandleFunc("/users", handlers.UsersGetHandler).Methods(http.MethodGet)

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
