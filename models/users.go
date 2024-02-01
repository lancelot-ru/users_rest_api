package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shakinm/xlsReader/xls"
	"github.com/xuri/excelize/v2"
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

	currentDate := pgtype.Date{Time: time.Now(), Valid: true}

	if _, err := GetDB().Exec(context.Background(),
		"INSERT INTO users (surname, name, patronymic, sex, status, date_of_birth, date_added) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.Surname, u.Name, u.Patronymic, u.Sex, u.Status, u.DateOfBirth, currentDate,
	); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"result": "user successfully created"})
}

func CreateUsersFromXLS(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer file.Close()

	f, err := xls.OpenReader(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	users := []User{}

	for _, sheet := range f.GetSheets() {
		for i := 0; i < sheet.GetNumberRows(); i++ {
			row, err := sheet.GetRow(i)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("row %d doesn't exist", i))
				return
			}

			var u User
			if len(row.GetCols()) != 5 {
				respondWithError(w, http.StatusInternalServerError, "invalid XLS file")
				return
			}

			cells := make([]string, 0, 5)
			for _, col := range row.GetCols() {
				cells = append(cells, col.GetString())
			}

			u.Surname = cells[0]
			u.Name = cells[1]
			u.Patronymic = cells[2]
			u.Sex = cells[3]
			u.Status = "active"
			u.DateOfBirth = pgtype.Date{}
			u.DateAdded = pgtype.Date{Time: time.Now(), Valid: true}

			users = append(users, u)
		}
	}

	var b strings.Builder
	b.WriteString("INSERT INTO users (surname, name, patronymic, sex, status, date_of_birth, date_added) VALUES")
	for _, u := range users {
		str := fmt.Sprintf(" ('%s', '%s',' %s', '%s', '%s', '%s', '%s'),", u.Surname, u.Name, u.Patronymic, u.Sex, u.Status, u.DateOfBirth.Time.Format("2006-01-02"), u.DateAdded.Time.Format("2006-01-02"))
		b.WriteString(str)
	}

	if _, err := GetDB().Exec(context.Background(), strings.Trim(b.String(), ",")); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"result": "user(s) successfully created"})
}

func CreateUsersFromXLSX(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer f.Close()

	users := []User{}

	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		for _, row := range rows {
			var u User
			if len(row) != 5 {
				respondWithError(w, http.StatusInternalServerError, "invalid XLSX file")
				return
			}

			u.Surname = row[0]
			u.Name = row[1]
			u.Patronymic = row[2]
			u.Sex = row[3]
			u.Status = "active"
			u.DateOfBirth = pgtype.Date{}
			u.DateAdded = pgtype.Date{Time: time.Now(), Valid: true}

			users = append(users, u)
		}
	}

	var b strings.Builder
	b.WriteString("INSERT INTO users (surname, name, patronymic, sex, status, date_of_birth, date_added) VALUES")
	for _, u := range users {
		str := fmt.Sprintf(" ('%s', '%s',' %s', '%s', '%s', '%s', '%s'),", u.Surname, u.Name, u.Patronymic, u.Sex, u.Status, u.DateOfBirth.Time.Format("2006-01-02"), u.DateAdded.Time.Format("2006-01-02"))
		b.WriteString(str)
	}

	if _, err := GetDB().Exec(context.Background(), strings.Trim(b.String(), ",")); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"result": "user(s) successfully created"})
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	row := GetDB().QueryRow(context.Background(), "SELECT * FROM users WHERE id=$1", id)

	var u User
	err = row.Scan(&u.Surname, &u.Name, &u.Patronymic, &u.Sex, &u.Status, &u.DateOfBirth, &u.DateAdded, &u.ID)
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

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if _, err := GetDB().Exec(context.Background(),
		"UPDATE users SET surname=$1, name=$2, patronymic=$3, sex=$4, status=$5, date_of_birth=$6 WHERE id=$7",
		u.Surname, u.Name, u.Patronymic, u.Sex, u.Status, u.DateOfBirth, id,
	); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "user successfully updated"})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if _, err := GetDB().Exec(context.Background(), "DELETE FROM users WHERE id=$1", id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "user successfully deleted"})
}
