package main

import (
	"context"
	"fmt"
	"log"
	"net/http" 
	"os"
	"time"
	"tours-service/db"
	"tours-service/internal/handlers"
	"tours-service/internal/repositories"
	"github.com/gorilla/mux"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := db.InitDB()
	if err != nil {
		log.Fatalf("Could not connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatalf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	toursDB := client.Database(os.Getenv("DB_NAME")) 

	tourRepo := repositories.NewTourRepository(toursDB)
	keypointRepo := repositories.NewKeypointRepository(toursDB)
	
	tourHandler := handlers.NewTourHandler(tourRepo, keypointRepo)
	keypointHandler := handlers.NewKeypointHandler(keypointRepo, tourRepo)

	router := mux.NewRouter()
	
	// Tour routes
	router.HandleFunc("/create", tourHandler.CreateTour).Methods("POST")
	router.HandleFunc("/my-tours", tourHandler.GetToursByAuthor).Methods("GET")
	router.HandleFunc("/create-tour-with-keypoints", tourHandler.CreateTourWithKeypoints).Methods("POST")
	
	// Keypoint routes
	router.HandleFunc("/{tourId}/addKeypoint", keypointHandler.CreateKeypoint).Methods("POST")
	router.HandleFunc("/{tourId}/keypoints", keypointHandler.GetKeypointsByTourID).Methods("GET")
	router.HandleFunc("/keypoints/{keypointId}", keypointHandler.GetKeypointByID).Methods("GET")
	router.HandleFunc("/keypoints/{keypointId}", keypointHandler.UpdateKeypoint).Methods("PUT")
	router.HandleFunc("/keypoints/{keypointId}", keypointHandler.DeleteKeypoint).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" 
	}

	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}