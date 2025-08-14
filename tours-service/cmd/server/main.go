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
	reviewRepo := repositories.NewTourReviewRepository(toursDB)
	
	
	tourHandler := handlers.NewTourHandler(tourRepo)
	keypointHandler := handlers.NewKeypointHandler(keypointRepo, tourRepo)
	reviewHandler := handlers.NewTourReviewHandler(reviewRepo, tourRepo)

	router := mux.NewRouter()
	
	// Tour routes
	router.HandleFunc("/api/create", tourHandler.CreateTour).Methods("POST")
	router.HandleFunc("/api/my-tours", tourHandler.GetToursByAuthor).Methods("GET")
	
	// Keypoint routes
	router.HandleFunc("/api/tours/{tourId}/addKeypoint", keypointHandler.CreateKeypoint).Methods("POST")
	router.HandleFunc("/api/tours/{tourId}/keypoints", keypointHandler.GetKeypointsByTourID).Methods("GET")
	router.HandleFunc("/api/keypoints/{keypointId}", keypointHandler.GetKeypointByID).Methods("GET")
	router.HandleFunc("/api/keypoints/{keypointId}", keypointHandler.UpdateKeypoint).Methods("PUT")
	router.HandleFunc("/api/keypoints/{keypointId}", keypointHandler.DeleteKeypoint).Methods("DELETE")

	// Review routes
	router.HandleFunc("/api/reviews", reviewHandler.CreateTourReview).Methods("POST")
	router.HandleFunc("/api/tours/{tourId}/reviews", reviewHandler.GetReviewsByTourID).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" 
	}

	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}