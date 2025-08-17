package main

import (
	"context"
	"fmt"
	"follower-service/data"
	"follower-service/handlers"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {

	port := os.Getenv("PORT")

	if len(port) == 0 {
		port = "8080"
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "[follower-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[follower-store] ", log.LstdFlags)

	store, err := data.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseDriverConnection(timeoutContext)
	store.CheckConnection()

	followerHandler := handlers.NewFollowerHandler(logger, store)

	router := mux.NewRouter()

	router.HandleFunc("/api/myFollowers", followerHandler.GetFollowers).Methods("GET")
	router.HandleFunc("/api/followedByMe", followerHandler.GetFollowed).Methods("GET")
	// router.HandleFunc("/api/suggestedFollowers", followerHandler.GetSuggested).Methods("GET")
	router.HandleFunc("/api/followedByMe/{id}", followerHandler.IsFollowedByMe).Methods("GET")
	router.HandleFunc("/api/followUser", followerHandler.Follow).Methods("POST")
	router.HandleFunc("/api/unfollowUser", followerHandler.Unfollow).Methods("POST")

	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
