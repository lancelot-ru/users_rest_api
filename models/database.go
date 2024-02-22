package models

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var db *pgx.Conn

var redisDb *redis.Client

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0, // default DB
	})

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to Redis: %v\n", err)
		os.Exit(1)
	}

	redisDb = client

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	db = conn
}

func GetDB() *pgx.Conn {
	return db
}

func GetRedisDB() *redis.Client {
	return redisDb
}

func FlushRedisDB() {
	err := GetRedisDB().FlushDB(context.Background()).Err()
	if err != nil {
		log.Println("Error flushing Redis DB")
	} else {
		log.Println("Successfully flushed Redis DB")
	}
}
