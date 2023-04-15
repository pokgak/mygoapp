package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Person struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var (
	DATABASE_FILE string = "database.json"
	db            *sql.DB
)

func main() {
	var err error

	// initialize database	
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatalln("Failed to open database file")
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS persons (
        id INTEGER PRIMARY KEY,
        name TEXT,
        age INTEGER
    )`)
	if err != nil {
		log.Fatalln("Failed to initialize database", err)
	}
	defer db.Close()

	// configure router
	r := mux.NewRouter()

	r.HandleFunc("/users", usersPostHandler).Methods(http.MethodPost)
	r.HandleFunc("/users", usersGetHandler).Methods(http.MethodGet)

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", r)
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

	id, err := addPerson(p, db)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}

	p.ID = id
	json.NewEncoder(w).Encode(p)
}

func usersGetHandler(w http.ResponseWriter, r *http.Request) {
	persons, err := getPersons(db)
	if err != nil {
		http.Error(w, "Failed to read database file", http.StatusInternalServerError)
		return
	}

	if len(persons) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = json.NewEncoder(w).Encode(persons)
	if err != nil {
		http.Error(w, "There was an error encoding the request body into the struct", http.StatusInternalServerError)
		return
	}
}

func addPerson(person Person, db *sql.DB) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO persons(name, age) VALUES(?, ?)")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(person.Name, person.Age)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	person.ID = id

	return id, nil
}

func getPersons(db *sql.DB) ([]Person, error) {
	rows, err := db.Query("SELECT id, name, age FROM persons")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var persons []Person

	for rows.Next() {
		var p Person
		err := rows.Scan(&p.ID, &p.Name, &p.Age)
		if err != nil {
			return nil, err
		}

		persons = append(persons, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return persons, nil
}
