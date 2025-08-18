package main

import (
	"context"
	"fmt"
	"follower-service/data"
	"follower-service/handlers"
	proto "follower-service/proto"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := log.New(os.Stdout, "[follower-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[follower-store] ", log.LstdFlags)

	port := os.Getenv("PORT")

	if len(port) == 0 {
		port = "8080"
	}
	conn, err := grpc.Dial("stakeholders-service:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close()

	followerClient := proto.NewStakeholdersServiceClient(conn)

	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	store, err := data.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseDriverConnection(timeoutContext)
	store.CheckConnection()

	followerHandler := handlers.NewFollowerHandler(logger, store, followerClient)

	router := mux.NewRouter()
	api := router.PathPrefix("/api/follow").Subrouter()

	api.HandleFunc("/myFollowers", followerHandler.GetFollowers).Methods("GET")
	api.HandleFunc("/followedByMe", followerHandler.GetFollowed).Methods("GET")
	api.HandleFunc("/suggestions/{limit}", followerHandler.GetSuggested).Methods("GET")
	api.HandleFunc("/followedByMe/{id}", followerHandler.IsFollowedByMe).Methods("GET")
	api.HandleFunc("/followUser", followerHandler.Follow).Methods("POST")
	api.HandleFunc("/unfollowUser", followerHandler.Unfollow).Methods("POST")

	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
