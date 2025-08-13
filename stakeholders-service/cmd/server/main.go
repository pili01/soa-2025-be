package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"stakeholders-service/db"
	"stakeholders-service/internal/handlers"
	"stakeholders-service/internal/repository"

	"github.com/gorilla/mux"
)

func main() {
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer database.Close()

	userRepo := repository.NewUserRepository(database)
	userHandler := handlers.NewUserHandler(userRepo)

	router := mux.NewRouter()
	router.HandleFunc("/api/register", userHandler.RegisterUser).Methods("POST")
	router.HandleFunc("/api/login", userHandler.LoginUser).Methods("POST")
	router.HandleFunc("/api/profile", userHandler.GetMyProfile).Methods("GET")
	router.HandleFunc("/api/updateProfile", userHandler.UpdateMyProfile).Methods("PUT")

	router.HandleFunc("/api/admin/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/api/admin/users/block", userHandler.BlockUser).Methods("PUT")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
