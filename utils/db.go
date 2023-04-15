package utils

import (
	"database/sql"
	"log"
)

var (
	Db            *sql.DB
	err			  error
)

func InitDatabase() {
	Db, err = sql.Open("sqlite3", "./data.Db")
	if err != nil {
		log.Fatalln("Failed to open database file")
	}

	_, err = Db.Exec(`CREATE TABLE IF NOT EXISTS persons (
        id INTEGER PRIMARY KEY,
        name TEXT,
        age INTEGER
    )`)
	if err != nil {
		log.Fatalln("Failed to initialize database", err)
	}
	defer Db.Close()
}