package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"users_rest_api/models"

	"github.com/gorilla/mux"
)

func main() {
	db := models.GetDB()
	redisDb := models.GetRedisDB()
	defer func() {
		log.Println("Closing DB...")
		db.Close(context.Background())
		redisDb.Close()
		log.Println("DB closed")
	}()

	r := mux.NewRouter()

	r.HandleFunc("/users", models.GetUsers).Methods("GET")
	r.HandleFunc("/users/new/json", models.CreateUser).Methods("POST")
	r.HandleFunc("/users/new/xls", models.CreateUsersFromXLS).Methods("POST")
	r.HandleFunc("/users/new/xlsx", models.CreateUsersFromXLSX).Methods("POST")

	r.HandleFunc("/users/{id}", models.GetUser).Methods("GET")
	r.HandleFunc("/users/{id}", models.UpdateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", models.DeleteUser).Methods("DELETE")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		log.Fatal(http.ListenAndServe(":8080", r))
	}()

	<-interrupt
	log.Println("Interrupt received...")
}
