package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

var userFields = getUserFields()

func getUserFields() []string {
	var field []string
	v := reflect.ValueOf(User{})
	for i := 0; i < v.Type().NumField(); i++ {
		field = append(field, v.Type().Field(i).Tag.Get("json"))
	}
	return field
}

func isStringInSlice(strSlice []string, s string) bool {
	for _, v := range strSlice {
		if v == s {
			return true
		}
	}

	return false
}

func parseSortQuery(sortBy string) (string, error) {
	splits := strings.Split(sortBy, ".")
	if len(splits) != 2 {
		return "", errors.New("invalid sortBy query")
	}

	field, order := splits[0], splits[1]
	if order != "desc" && order != "asc" {
		return "", errors.New("invalid order of sortBy query")
	}

	if !isStringInSlice(userFields, field) {
		return "", errors.New("invalid field")
	}

	return fmt.Sprintf("%s %s", field, strings.ToUpper(order)), nil
}

func parseFilterMap(filter string) (map[string]string, error) {
	splits := strings.Split(filter, ".")
	if len(splits) != 2 {
		return nil, errors.New("invalid filter query")
	}

	field, value := splits[0], splits[1]

	filters := make(map[string]string)
	if field == "fullname" {
		splits = strings.Split(strings.Trim(value, "\""), " ")
		if len(splits) != 2 && len(splits) != 3 {
			return nil, errors.New("invalid filter query")
		}

		filters["surname"] = splits[0]
		filters["name"] = splits[1]
		if len(splits) == 3 {
			filters["patronymic"] = splits[2]
		}
		return filters, nil

	} else if !isStringInSlice(userFields, field) {
		return nil, errors.New("invalid field")
	}

	filters[field] = value
	return filters, nil
}

func constructQuery(filterMap map[string]string, sortQuery string, limit int, offset int) string {
	var b strings.Builder
	b.WriteString("SELECT * FROM users")

	if len(filterMap) != 0 {
		b.WriteString(" WHERE ")
		first := true
		for k, v := range filterMap {
			if !first {
				b.WriteString(" AND ")
			} else {
				first = false
			}
			b.WriteString(k + "='" + v + "'")
		}
	}

	if len(sortQuery) != 0 {
		b.WriteString(" ORDER BY ")
		b.WriteString(sortQuery)
	}
	if limit != -1 {
		b.WriteString(" LIMIT ")
		b.WriteString(strconv.Itoa(limit))
	}
	if offset != -1 {
		b.WriteString(" OFFSET ")
		b.WriteString(strconv.Itoa(offset))
	}

	return b.String()
}
