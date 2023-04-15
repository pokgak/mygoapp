package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/pokgak/mygoapp/models"
	"github.com/pokgak/mygoapp/utils"
)

func UsersPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer(utils.Tracer).Start(context.Background(), "usersPostHandler")
	defer span.End()

	var p models.Person

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	id, err := addPerson(ctx, p, utils.Db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	p.ID = id
	json.NewEncoder(w).Encode(p)
}

func UsersGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer(utils.Tracer).Start(context.Background(), "usersGetHandler")
	defer span.End()

	persons, err := getPersons(ctx, utils.Db)
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

func addPerson(ctx context.Context, person models.Person, db *sql.DB) (int64, error) {
	_, span := otel.Tracer(utils.Tracer).Start(ctx, "addPerson")
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

func getPersons(ctx context.Context, db *sql.DB) ([]models.Person, error) {
	_, span := otel.Tracer(utils.Tracer).Start(ctx, "getPersons")
	defer span.End()

	rows, err := db.Query("SELECT id, name, age FROM persons")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	defer rows.Close()

	var persons []models.Person

	for rows.Next() {
		var p models.Person
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
