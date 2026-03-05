package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	if sslRootCert := os.Getenv("DB_SSLROOTCERT"); sslRootCert != "" {
		dsn = fmt.Sprintf("%s sslrootcert=%s", dsn, sslRootCert)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	// Configure high-performance connection pool for 5000+ req/sec
	db.SetMaxOpenConns(100)                 // Increase from 25 to 100
	db.SetMaxIdleConns(50)                  // Keep 50 idle connections alive
	db.SetConnMaxIdleTime(10 * time.Minute) // Close idle connections after 10 min
	db.SetConnMaxLifetime(1 * time.Hour)    // Close connections older than 1 hour

	err = db.Ping()
	if err != nil {
		log.Fatalf("DB ping error: %v", err)
	}

	DB = db
	log.Println("Connected to Postgres with connection pooling")
}
