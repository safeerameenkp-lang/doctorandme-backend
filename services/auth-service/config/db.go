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
		log.Printf("Using SSL root cert: %s", sslRootCert)
	}

	// For debugging (password masked)
	maskedDSN := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s sslrootcert=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"), os.Getenv("DB_SSLMODE"), os.Getenv("DB_SSLROOTCERT"))
	log.Printf("Connecting to DB with DSN: %s", maskedDSN)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	// Configure high-performance connection pool for 5000+ req/sec
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
	db.SetConnMaxIdleTime(10 * time.Minute)
	db.SetConnMaxLifetime(1 * time.Hour)

	// Retry logic for DB ping
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for DB... (%d/10) error: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Final DB ping failure: %v", err)
	}

	DB = db
	log.Println("Connected to Postgres with connection pooling")
}
