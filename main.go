package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/go-playground/validator/v10"
)

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Price       int    `json:"price" validate:"required,min=0"`
}

func createProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)

	var product Product
	err := decoder.Decode(&product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Received product data: %v\n", product)

	validate := validator.New()
	err = validate.Struct(product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save the product to the database or perform other actions here...

	fmt.Fprintf(w, "Created product %s", product.Name)
}

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
	r.HandleFunc("/products/{productId}", productHandler).Methods(http.MethodGet)
	r.HandleFunc("/products", createProductHandler).Methods(http.MethodPost)

	fmt.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
