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

	saga "example.com/common/saga/messaging"
	natsmsg "example.com/common/saga/messaging/nats"

	pb "tours-service/proto/compiled"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

type config struct {
	NatsHost string
	NatsPort string
	NatsUser string
	NatsPass string

	CommandSubject string
	ReplySubject   string
	QueueGroup     string
}

func loadConfig() config {
	c := config{
		NatsHost:       getenv("NATS_HOST", "localhost"),
		NatsPort:       getenv("NATS_PORT", "4222"),
		NatsUser:       getenv("NATS_USER", ""),
		NatsPass:       getenv("NATS_PASS", ""),
		CommandSubject: getenv("PURCHASE_COMMAND_SUBJECT", "purchase.checkout.command"),
		ReplySubject:   getenv("PURCHASE_REPLY_SUBJECT", "purchase.checkout.reply"),
		QueueGroup:     getenv("NATS_QUEUE_GROUP", "purchase_service"),
	}
	return c
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustPublisher(cfg config) saga.Publisher {

	pub, err := natsmsg.NewNATSPublisher(cfg.NatsHost, cfg.NatsPort, cfg.NatsUser, cfg.NatsPass, cfg.ReplySubject)
	if err != nil {
		log.Fatal("greska: tours publisher")
	}
	return pub
}

func mustSubscriber(cfg config) saga.Subscriber {
	sub, err := natsmsg.NewNATSSubscriber(cfg.NatsHost, cfg.NatsPort, cfg.NatsUser, cfg.NatsPass, cfg.CommandSubject, "tours_handler_group")
	if err != nil {
		log.Fatal("greska: tours subscriber")
	}
	return sub
}

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
	reviewRepo := repositories.NewTourReviewRepository(toursDB)
	tourExecutionRepo := repositories.NewTourExecutionRepository(toursDB)
	capacityRepo := repositories.NewTourCapacityRepository(toursDB)

	// --- Services ---
	mapService := services.NewMapService(os.Getenv("MAP_SERVICE_URL"))
	tourService := services.NewTourService(tourRepo, keypointRepo, mapService)
	tourReviewService := services.NewTourReviewService(reviewRepo)
	keypointService := services.NewKeypointService(keypointRepo)
	authService := services.NewAuthService()
	purchaseService := services.NewPurchaseService()
	tourExecutionService := services.NewTourExecutionService(tourExecutionRepo, tourService, keypointService)
	capacityService := services.NewTourCapacityService(capacityRepo)

	// --- HTTP Handlers ---
	tourHandler := handlers.NewTourHandler(tourService, keypointService, tourReviewService, authService)
	keypointHandler := handlers.NewKeypointHandler(keypointService, tourService, authService)
	reviewHandler := handlers.NewTourReviewHandler(tourReviewService, tourService, authService)
	TourExecutionHandler := handlers.NewTourExecutionHandler(tourExecutionService, authService, purchaseService)
	capHandler := handlers.NewCapacityHandler(capacityService, tourService, authService)

	log.Println("Initializing NATS and Saga handler...")

	natsCfg := loadConfig()
	replyPublisher := mustPublisher(natsCfg)
	commandSubscriber := mustSubscriber(natsCfg)

	_, err = handlers.NewCreatePurchaseCommandHandler(capacityService, replyPublisher, commandSubscriber)
	if err != nil {
		log.Fatalf("Failed to create and subscribe purchase command handler: %v", err)
	}

	log.Println("Saga handler for purchase commands is running.")

	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	// --- Review routes ---
	api.HandleFunc("/reviews", reviewHandler.CreateTourReview).Methods("POST")
	api.HandleFunc("/{tourId}/reviews", reviewHandler.GetReviewsByTourID).Methods("GET")

	// --- Tour routes ---
	api.HandleFunc("/create", tourHandler.CreateTour).Methods("POST")
	api.HandleFunc("/my-tours", tourHandler.GetToursByAuthor).Methods("GET")
	api.HandleFunc("/get-published", tourHandler.GetPublishedToursWithFirstKeypoint).Methods("GET")
	api.HandleFunc("/{tourId}", tourHandler.GetTourByID).Methods("GET")
	api.HandleFunc("/{tourId}", tourHandler.UpdateTour).Methods("PUT")
	api.HandleFunc("/{tourId}", tourHandler.DeleteTour).Methods("DELETE")
	api.HandleFunc("/{tourId}/publish", tourHandler.PublishTour).Methods("POST")
	api.HandleFunc("/{tourId}/archive", tourHandler.ArchiveTour).Methods("POST")
	api.HandleFunc("/{tourId}/set-price", tourHandler.SetTourPrice).Methods("POST")
	api.HandleFunc("/{tourId}/tourist-view", tourHandler.GetTourForTourist).Methods("GET")
	api.HandleFunc("/{tourId}/purchased-keypoints", tourHandler.GetPurchasedKeypoints).Methods("GET")

	// --- Keypoint routes ---
	api.HandleFunc("/{tourId}/create-keypoint", keypointHandler.CreateKeypoint).Methods("POST")
	api.HandleFunc("/{tourId}/keypoints", keypointHandler.GetKeypointsByTourID).Methods("GET")
	api.HandleFunc("/keypoints/{keypointId}", keypointHandler.GetKeypointByID).Methods("GET")
	api.HandleFunc("/keypoints/{keypointId}", keypointHandler.UpdateKeypoint).Methods("PUT")
	api.HandleFunc("/keypoints/{keypointId}", keypointHandler.DeleteKeypoint).Methods("DELETE")
	api.HandleFunc("/keypoints/{keypointId}/upload-image", keypointHandler.UploadKeypointImage).Methods("POST")

	// -- Execution routes --
	executionRouter := api.PathPrefix("/execution").Subrouter()
	executionRouter.HandleFunc("/my-executions", TourExecutionHandler.GetExecutionsByUser).Methods("GET")
	executionRouter.HandleFunc("/tour/{tour_id}", TourExecutionHandler.GetMyExecutionByTourID).Methods("GET")
	executionRouter.HandleFunc("/start/{tour_id}", TourExecutionHandler.StartTourExecution).Methods("POST")
	executionRouter.HandleFunc("/abort/{tour_id}", TourExecutionHandler.AbortExecution).Methods("POST")
	executionRouter.HandleFunc("/is-keypoint-reached/{tour_id}", TourExecutionHandler.CheckIsKeyPointReached).Methods("POST")

	// Tour capacity
	api.HandleFunc("/capacity/{tourId}", capHandler.GetCapacity).Methods("GET")
	api.HandleFunc("/capacity/{tourId}", capHandler.InitOrUpdateCapacity).Methods("PUT")
	api.HandleFunc("/capacity/{tourId}/consume", capHandler.ConsumeSeats).Methods("POST")
	api.HandleFunc("/capacity/{tourId}/release", capHandler.ReleaseSeats).Methods("POST")

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
