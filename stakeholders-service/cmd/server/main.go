package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	_ "github.com/lib/pq" 
)

func main() {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=stakeholders-db port=5432 user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open a connection to the database:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	fmt.Println("Successfully connected to the database!")
}