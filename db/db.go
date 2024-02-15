package db

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sqlx.DB
}

func GetDB(secrets map[string]string) *PostgresDB {
	db := createDB(secrets)
	return &PostgresDB{
		DB: db,
	}
}

func createDB(secrets map[string]string) *sqlx.DB {

	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
		return nil
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s",
		dbHost, dbPort, secrets["username"], secrets["password"])

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		log.Fatal(err.Error())
	}
	// Create customers table
	createTableSQL := `
	CREATE TABLE customers (
	    id SERIAL PRIMARY KEY,
	    name VARCHAR(255),
	    email VARCHAR(255) UNIQUE,
	    address VARCHAR(255)
	);
	`
	_, err = db.Exec(createTableSQL)
	return db
}
