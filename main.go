package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var (
	DATABASE_FILE = "database.json"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/users", usersPostHandler).Methods(http.MethodPost)
	r.HandleFunc("/users", usersGetHandler).Methods(http.MethodGet)

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}

func usersPostHandler(w http.ResponseWriter, r *http.Request) {
	var p Person

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "There was an error decoding the request body into the struct", http.StatusInternalServerError)
		return
	}

	// Check if the database file exists
	_, err = os.Stat(DATABASE_FILE)
	if os.IsNotExist(err) {
		// If it doesn't exist, create it and write the object to it
		writeJSONToFile([]Person{p})
	} else {
		// If it exists, read its contents, append the new object, and write it back to the file
		persons, err := readJSONFromFile()
		if err != nil {
			log.Fatal(err)
		}
		persons = append(persons, p)
		writeJSONToFile(persons)
	}
}

func usersGetHandler(w http.ResponseWriter, r *http.Request) {
	_, err := os.Stat(DATABASE_FILE)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	persons, err := readJSONFromFile()
	if err != nil {
		http.Error(w, "Failed to read database file", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(persons)
	if err != nil {
		http.Error(w, "There was an error encoding the request body into the struct", http.StatusInternalServerError)
		return
	}
}

func readJSONFromFile() ([]Person, error) {
	var persons []Person
	file, err := os.Open(DATABASE_FILE)
	if err != nil {
		return persons, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&persons)
	if err != nil {
		return persons, err
	}
	return persons, nil
}

func writeJSONToFile(persons []Person) {
	file, err := os.Create(DATABASE_FILE)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(persons)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Object added to database.json")
}
