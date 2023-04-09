package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to the home page!")
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	productId := mux.Vars(r)["productId"]
	fmt.Fprintf(w, "You requested product %s", productId)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/products/{productId}", productHandler)

	fmt.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
