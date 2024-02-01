package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	Surname     string      `json:"surname"`
	Name        string      `json:"name"`
	Patronymic  string      `json:"patronymic,omitempty"`
	Sex         string      `json:"sex"`
	Status      string      `json:"status"`
	DateOfBirth pgtype.Date `json:"date_of_birth,omitempty"`
	DateAdded   pgtype.Date `json:"date_added"`
	ID          int         `json:"id"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var u User
	_ = json.NewDecoder(r.Body).Decode(&u)

	if _, err := GetDB().Exec(context.Background(),
		"INSERT INTO users (surname, name, patronymic, sex, status, date_of_birth, date_added) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.Surname, u.Name, u.Patronymic, u.Sex, u.Status, u.DateOfBirth, u.DateAdded,
	); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func CreateUserFromXLS(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusInternalServerError, "TODO")
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	row := GetDB().QueryRow(context.Background(), "SELECT * FROM users WHERE id=$1", params["id"])

	var u User
	err := row.Scan(&u.Surname, &u.Name, &u.Patronymic, &u.Sex, &u.Status, &u.DateOfBirth, &u.DateAdded, &u.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	users := []User{}

	sortBy := r.URL.Query().Get("sortBy")
	if sortBy == "" {
		sortBy = "id.asc"
	}

	sortQuery, err := parseSortQuery(sortBy)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	strLimit := r.URL.Query().Get("limit")
	limit := -1
	if strLimit != "" {
		limit, err = strconv.Atoi(strLimit)
		if err != nil || limit < -1 {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	strOffset := r.URL.Query().Get("offset")
	offset := -1
	if strOffset != "" {
		offset, err = strconv.Atoi(strOffset)
		if err != nil || offset < -1 {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	filter := r.URL.Query().Get("filter")
	filterMap := map[string]string{}
	if filter != "" {
		filterMap, err = parseFilterMap(filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	queryString := constructQuery(filterMap, sortQuery, limit, offset)
	fmt.Println(queryString)
	rows, err := GetDB().Query(context.Background(), queryString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		err := rows.Scan(&u.Surname, &u.Name, &u.Patronymic, &u.Sex, &u.Status, &u.DateOfBirth, &u.DateAdded, &u.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		users = append(users, u)
	}

	err = rows.Err()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, users)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var u User
	_ = json.NewDecoder(r.Body).Decode(&u)

	params := mux.Vars(r)

	if _, err := GetDB().Exec(context.Background(),
		"UPDATE users SET surname=$1, name=$2, patronymic=$3, sex=$4, status=$5, date_of_birth=$6, date_added=$7 WHERE id=$8",
		u.Surname, u.Name, u.Patronymic, u.Sex, u.Status, u.DateOfBirth, u.DateAdded, params["id"],
	); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if _, err := GetDB().Exec(context.Background(), "DELETE FROM users WHERE id=$1", params["id"]); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
