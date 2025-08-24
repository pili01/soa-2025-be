package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"tours-service/db"
	"tours-service/internal/grpc_handlers"
	"tours-service/internal/handlers"
	"tours-service/internal/repositories"
	"tours-service/internal/services"

	pb "tours-service/proto/compiled"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// --- MongoDB init ---
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

	// --- Repositories ---
	tourRepo := repositories.NewTourRepository(toursDB)
	keypointRepo := repositories.NewKeypointRepository(toursDB)
	tourExecutionRepo := repositories.NewTourExecutionRepository(toursDB)

	// --- Services ---
	mapService := services.NewMapService(os.Getenv("MAP_SERVICE_URL"))
	tourService := services.NewTourService(tourRepo, keypointRepo, mapService)
	keypointService := services.NewKeypointService(keypointRepo)
	authService := services.NewAuthService()
	tourExecutionService := services.NewTourExecutionService(tourExecutionRepo, tourService, keypointService)

	// --- HTTP Handlers ---
	tourHandler := handlers.NewTourHandler(tourService, authService)
	keypointHandler := handlers.NewKeypointHandler(keypointService, tourService, authService)
	TourExecutionHandler := handlers.NewTourExecutionHandler(tourExecutionService, authService)

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()

	// --- Tour routes ---
	apiRouter.HandleFunc("/create", tourHandler.CreateTour).Methods("POST")
	apiRouter.HandleFunc("/my-tours", tourHandler.GetToursByAuthor).Methods("GET")
	apiRouter.HandleFunc("/{tourId}", tourHandler.GetTourByID).Methods("GET")
	apiRouter.HandleFunc("/{tourId}", tourHandler.UpdateTour).Methods("PUT")
	apiRouter.HandleFunc("/{tourId}", tourHandler.DeleteTour).Methods("DELETE")
	apiRouter.HandleFunc("/{tourId}/publish", tourHandler.PublishTour).Methods("POST")
	apiRouter.HandleFunc("/{tourId}/archive", tourHandler.ArchiveTour).Methods("POST")
	apiRouter.HandleFunc("/{tourId}/set-price", tourHandler.SetTourPrice).Methods("POST")
	apiRouter.HandleFunc("/get-published", tourHandler.GetPublishedToursWithFirstKeypoint).Methods("GET")

	// --- Keypoint routes ---
	apiRouter.HandleFunc("/{tourId}/create-keypoint", keypointHandler.CreateKeypoint).Methods("POST")
	apiRouter.HandleFunc("/{tourId}/keypoints", keypointHandler.GetKeypointsByTourID).Methods("GET")
	apiRouter.HandleFunc("/keypoints/{keypointId}", keypointHandler.GetKeypointByID).Methods("GET")
	apiRouter.HandleFunc("/keypoints/{keypointId}", keypointHandler.UpdateKeypoint).Methods("PUT")
	apiRouter.HandleFunc("/keypoints/{keypointId}", keypointHandler.DeleteKeypoint).Methods("DELETE")

	// -- Execution routes --
	executionRouter := apiRouter.PathPrefix("/execution").Subrouter()
	executionRouter.HandleFunc("/start/{tour_id}", TourExecutionHandler.StartTourExecution).Methods("POST")
	executionRouter.HandleFunc("/abort/{tour_id}", TourExecutionHandler.AbortExecution).Methods("POST")
	executionRouter.HandleFunc("/is-keypoint-reached/{tour_id}", TourExecutionHandler.CheckIsKeyPointReached).Methods("POST")

	// --- Start gRPC Server ---
	grpcLis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()
	tourGRPCServer := grpc_handlers.NewTourGRPCServer(tourService)
	pb.RegisterTourServiceServer(grpcServer, tourGRPCServer)

	go func() {
		log.Println("gRPC server listening on :50051")
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// --- Start HTTP Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("HTTP server running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
