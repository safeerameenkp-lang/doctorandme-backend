package utils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
)

// GenerateClinicCode retrieves an incremental clinic code based on the clinic name using DB transaction locks
func GenerateClinicCode(ctx context.Context, tx *sql.Tx, db *sql.DB, name string) (string, error) {
	// Extract first letter of each word (Uppercase)
	words := strings.Fields(name)
	prefix := ""
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")

	for _, word := range words {
		cleanWord := reg.ReplaceAllString(word, "")
		if len(cleanWord) > 0 {
			prefix += string(cleanWord[0])
		}
	}
	prefix = strings.ToUpper(prefix)

	// Ensure prefix length rules
	if len(prefix) == 0 {
		return "", errors.New("cannot generate clinic code from given name")
	}

	// 1. Transactional Postgres Advisory Lock for High Concurrency (1000+ req/sec)
	// Hash prefix to 64-bit int for locking specifically this prefix globally
	h := fnv.New64a()
	h.Write([]byte("CLINIC_" + prefix))
	lockID := int64(h.Sum64() & 0x7FFFFFFFFFFFFFFF) // Safe positive bounds for PG

	if tx != nil {
		_, _ = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, lockID)
	}

	// 2. Fetch the highest suffix mapped
	var maxNumber int
	query := `
		SELECT COALESCE(MAX(
			CASE 
				WHEN clinic_code ~ ('^' || $1 || '[0-9]+$') 
				THEN CAST(SUBSTRING(clinic_code FROM LENGTH($1) + 1) AS INTEGER)
				ELSE 0
			END
		), 0) as max_num
		FROM clinics 
	`

	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, prefix).Scan(&maxNumber)
	} else {
		// If tx is nil, advisory lock cannot be maintained across boundaries
		err = db.QueryRowContext(ctx, query, prefix).Scan(&maxNumber)
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%03d", prefix, maxNumber+1), nil
}

// GenerateDoctorCode creates unique initials-based code incrementally
func GenerateDoctorCode(ctx context.Context, tx *sql.Tx, db *sql.DB, firstName, lastName string) (string, error) {
	// Remove "Dr" prefix
	cleanFirst := strings.TrimSpace(firstName)
	if strings.HasPrefix(strings.ToLower(cleanFirst), "dr ") || strings.HasPrefix(strings.ToLower(cleanFirst), "dr. ") {
		parts := strings.SplitN(cleanFirst, " ", 2)
		if len(parts) > 1 {
			cleanFirst = strings.TrimSpace(parts[1])
		} else {
			cleanFirst = ""
		}
	}

	fullName := strings.TrimSpace(cleanFirst + " " + lastName)

	words := strings.Fields(fullName)
	prefix := ""
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")

	for _, word := range words {
		cleanWord := reg.ReplaceAllString(word, "")
		if len(cleanWord) > 0 {
			prefix += string(cleanWord[0])
		}
	}
	prefix = strings.ToUpper(prefix)
	if len(prefix) == 0 {
		return "", errors.New("cannot generate doctor code")
	}

	// Concurrency lock using Postgres Advisory Locks matching Doctor Prefix
	h := fnv.New64a()
	h.Write([]byte("DOCTOR_" + prefix))
	lockID := int64(h.Sum64() & 0x7FFFFFFFFFFFFFFF)

	if tx != nil {
		_, _ = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, lockID)
	}

	var maxNumber int
	query := `
		SELECT COALESCE(MAX(
			CASE 
				WHEN doctor_code ~ ('^' || $1 || '[0-9]+$') 
				THEN CAST(SUBSTRING(doctor_code FROM LENGTH($1) + 1) AS INTEGER)
				ELSE 0
			END
		), 0) as max_num
		FROM doctors
	`

	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, prefix).Scan(&maxNumber)
	} else {
		err = db.QueryRowContext(ctx, query, prefix).Scan(&maxNumber)
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%03d", prefix, maxNumber+1), nil
}
