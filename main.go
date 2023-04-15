package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
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

const (
	tracer = "mygoapp"
)

// newExporter returns a console exporter.
func newExporter(w io.Writer) (trace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

// newResource returns a resource describing this application.
func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("mygoapp"),
			semconv.ServiceVersion("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	return r
}

func main() {
	var err error

	// Write telemetry data to a file.
	f, err := os.Create("traces.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	exp, err := newExporter(f)
	if err != nil {
		log.Fatal(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(newResource()),
	)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()
	otel.SetTracerProvider(tp)

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
	ctx, span := otel.Tracer(tracer).Start(context.Background(), "usersPostHandler")
	defer span.End()

	var p Person

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	id, err := addPerson(ctx, p, db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	p.ID = id
	json.NewEncoder(w).Encode(p)
}

func usersGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer(tracer).Start(context.Background(), "usersGetHandler")
	defer span.End()

	persons, err := getPersons(ctx, db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	span.SetAttributes(attribute.Int("users.count", len(persons)))

	if len(persons) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = json.NewEncoder(w).Encode(persons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

}

func addPerson(ctx context.Context, person Person, db *sql.DB) (int64, error) {
	_, span := otel.Tracer(tracer).Start(ctx, "addPerson")
	defer span.End()

	stmt, err := db.Prepare("INSERT INTO persons(name, age) VALUES(?, ?)")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(person.Name, person.Age)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	span.SetAttributes(attribute.Int64("user.id", id))

	person.ID = id

	return id, nil
}

func getPersons(ctx context.Context, db *sql.DB) ([]Person, error) {
	_, span := otel.Tracer(tracer).Start(ctx, "getPersons")
	defer span.End()

	rows, err := db.Query("SELECT id, name, age FROM persons")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	defer rows.Close()

	var persons []Person

	for rows.Next() {
		var p Person
		err := rows.Scan(&p.ID, &p.Name, &p.Age)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}

		persons = append(persons, p)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return persons, nil
}
