package driver

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
    connStr := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
    )

    fmt.Println("Waiting for database....")
    time.Sleep(5 * time.Second)

    var err error
    db, err = sql.Open("postgres", connStr) // ⬅️ fix: jangan shadowing
    if err != nil {
        log.Fatalf("Failed to open DB: %v", err)
    }

    // optional: retry loop supaya lebih tahan
    for i := 0; i < 5; i++ {
        err = db.Ping()
        if err == nil {
            fmt.Println("Successfully connected to the database")
            return
        }
        log.Println("DB not ready, retrying...")
        time.Sleep(2 * time.Second)
    }

    log.Fatalf("Failed to ping DB: %v", err)
}

func GetDB() *sql.DB {
	return db
}

func CloseDB() {
	if err := db.Close(); err != nil {
		log.Fatalf("Failed to close the database connection: %v", err)
	}
}