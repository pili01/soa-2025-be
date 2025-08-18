package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"stakeholders-service/db"
	"stakeholders-service/internal/handlers"
	repository "stakeholders-service/internal/repositories"
	proto "stakeholders-service/proto"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
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
	router.HandleFunc("/api/validateRole", userHandler.ValidateRole).Methods("POST")
	router.HandleFunc("/api/updateProfile", userHandler.UpdateMyProfile).Methods("PUT")
	router.HandleFunc("/api/me", userHandler.GetUserFromToken).Methods("GET")

	router.HandleFunc("/api/admin/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/api/admin/users/block", userHandler.BlockUser).Methods("PUT")

	go func() {
		lis, err := net.Listen("tcp", ":8000")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		proto.RegisterStakeholdersServiceServer(grpcServer, userHandler)
		fmt.Println("gRPC server started on port 8000")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
