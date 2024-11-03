package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func connectionString() string {
	var url string = "postgresql://"
	url += os.Getenv("DB_USER") + ":"
	url += os.Getenv("DB_PSWD") + "@"
	url += os.Getenv("DB_HOST") + ":"
	url += os.Getenv("DB_PORT") + "/"
	url += os.Getenv("DB_DATABASE") + "?sslmode=verify-full"
	return url
}

func Connect() *sql.DB {
	db, err := sql.Open("postgres", connectionString())
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func AddEvent(db *sql.DB, event map[string]string) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("> failed to marshal event data...")
	}

	sql := `INSERT INTO %s.events.sink (event_time, event) VALUES($1, $2);`
	stmt, err := db.Prepare(fmt.Sprintf(sql, os.Getenv("DB_DATABASE")))
	if err != nil {
		log.Printf("> failed to prepare insert: %v", err)
	}

	result, err := stmt.Exec(time.Now().Unix(), eventJSON)
	if err != nil {
		log.Printf("db-err: %s\n", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 1 {
		log.Printf("successfully added event: {%s}", eventJSON)
	} else {
		log.Printf("failed to add event: {%s}", eventJSON)
	}
}
