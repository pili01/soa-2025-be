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
	"tours-service/internal/services"

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

	mapService := services.NewMapService(os.Getenv("MAP_SERVICE_URL"))
	tourService := services.NewTourService(tourRepo, keypointRepo, mapService)
	keypointService := services.NewKeypointService(keypointRepo)
	authService := services.NewAuthService()

	// Handlers
	tourHandler := handlers.NewTourHandler(tourService, authService)
	keypointHandler := handlers.NewKeypointHandler(keypointService, tourService, authService)
	reviewHandler := handlers.NewTourReviewHandler(reviewRepo, tourRepo)

	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	// Review routes
	api.HandleFunc("/reviews", reviewHandler.CreateTourReview).Methods("POST")
	api.HandleFunc("/tours/{tourId}/reviews", reviewHandler.GetReviewsByTourID).Methods("GET")

	// Tour routes
	api.HandleFunc("/create", tourHandler.CreateTour).Methods("POST")
	api.HandleFunc("/my-tours", tourHandler.GetToursByAuthor).Methods("GET")
	api.HandleFunc("/{tourId}", tourHandler.GetTourByID).Methods("GET")
	api.HandleFunc("/{tourId}", tourHandler.UpdateTour).Methods("PUT")
	api.HandleFunc("/{tourId}", tourHandler.DeleteTour).Methods("DELETE")
	api.HandleFunc("/{tourId}/publish", tourHandler.PublishTour).Methods("POST")
	api.HandleFunc("/{tourId}/archive", tourHandler.ArchiveTour).Methods("POST")
	api.HandleFunc("/{tourId}/set-price", tourHandler.SetTourPrice).Methods("POST")
	api.HandleFunc("/get-published", tourHandler.GetPublishedToursWithFirstKeypoint).Methods("GET")

	// Keypoint routes
	api.HandleFunc("/{tourId}/create-keypoint", keypointHandler.CreateKeypoint).Methods("POST")
	api.HandleFunc("/{tourId}/keypoints", keypointHandler.GetKeypointsByTourID).Methods("GET")
	api.HandleFunc("/keypoints/{keypointId}", keypointHandler.GetKeypointByID).Methods("GET")
	api.HandleFunc("/keypoints/{keypointId}", keypointHandler.UpdateKeypoint).Methods("PUT")
	api.HandleFunc("/keypoints/{keypointId}", keypointHandler.DeleteKeypoint).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
