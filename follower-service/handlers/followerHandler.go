package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"follower-service/data"
	"log"
	"net/http"
	"strconv"

	proto "follower-service/proto" // Add this import for proto

	"github.com/gorilla/mux"
	"google.golang.org/grpc/metadata"
)

type FollowerHandler struct {
	logger                *log.Logger
	repo                  *data.FollowerRepo
	followerServiceClient proto.StakeholdersServiceClient
}

func NewFollowerHandler(l *log.Logger, r *data.FollowerRepo, f proto.StakeholdersServiceClient) *FollowerHandler {
	return &FollowerHandler{
		logger:                l,
		repo:                  r,
		followerServiceClient: f,
	}
}

func (fh *FollowerHandler) getAuthorization(rw http.ResponseWriter, h *http.Request) (int, string, string, error) {
	authHeader := h.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(rw, "Authorization header is missing", http.StatusUnauthorized)
		return 0, "", "", fmt.Errorf("authorization header is missing")
	}
	md := metadata.New(map[string]string{"authorization": authHeader})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	response, err := fh.followerServiceClient.GetMyInfo(ctx, &proto.GetMyInfoRequest{})
	if err != nil {
		http.Error(rw, "Failed to get user info", http.StatusForbidden)
		return 0, "", "", fmt.Errorf("failed to get user info: %w", err)
	}
	userID := response.Id
	username := response.Username
	role := response.Role
	return (int)(userID), username, role, nil
}

func (fh *FollowerHandler) GetFollowers(rw http.ResponseWriter, h *http.Request) {
	userID, username, role, err := fh.getAuthorization(rw, h)
	fmt.Printf("Getting followers for user ID: %d, Username: %s, Role: %s\n", userID, username, role)
	if err != nil {
		return
	}
	if role != "Tourist" && role != "Guide" {
		http.Error(rw, "Access denied", http.StatusForbidden)
		return
	}
	fmt.Println("Getting followers...")
	followers, err := fh.repo.GetFollowers(userID)
	if err != nil {
		fh.logger.Print("Database exception: ", err)
		http.Error(rw, "Database error", http.StatusInternalServerError)
		return
	}
	if len(followers) == 0 {
		rw.WriteHeader(http.StatusOK)
		emptyList := []data.User{}
		json.NewEncoder(rw).Encode(emptyList)
		return
	}

	for _, follower := range followers {
		isFollowed, err := fh.repo.IsFollowedByMe(userID, follower.ID)
		if err != nil {
			fh.logger.Print("Database exception: ", err)
			http.Error(rw, "Database error", http.StatusInternalServerError)
			return
		}
		if len(isFollowed) == 0 {
			follower.FollowedByMe = false
			continue
		}
		follower.FollowedByMe = true
	}
	err = followers.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		fh.logger.Print("Unable to convert to json :", err)
		return
	}
}

func (fh *FollowerHandler) IsFollowedByMe(rw http.ResponseWriter, h *http.Request) {
	userID, username, role, err := fh.getAuthorization(rw, h)
	fmt.Printf("Checking following for user ID: %d, Username: %s, Role: %s\n", userID, username, role)
	if err != nil {
		return
	}
	if role != "Tourist" && role != "Guide" {
		http.Error(rw, "Access denied", http.StatusForbidden)
		return
	}
	vars := mux.Vars(h)
	followedIdStr := vars["id"]
	fmt.Printf("Checking if followed by me for user id: %s\n", followedIdStr)
	followedId, err := strconv.Atoi(followedIdStr)
	if err != nil {
		fh.logger.Print("Invalid id parameter: ", err)
		http.Error(rw, "Invalid id parameter", http.StatusBadRequest)
		return
	}
	followers, err := fh.repo.IsFollowedByMe(userID, followedId)
	if err != nil {
		fh.logger.Print("Database exception: ", err)
		http.Error(rw, "Database error", http.StatusInternalServerError)
		return
	}
	if len(followers) == 0 {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{"value": false}`))
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte(`{"value": true}`))
}

func (fh *FollowerHandler) GetFollowed(rw http.ResponseWriter, h *http.Request) {
	userID, username, role, err := fh.getAuthorization(rw, h)
	fmt.Printf("Getting followed users for user ID: %d, Username: %s, Role: %s\n", userID, username, role)
	if err != nil {
		return
	}
	if role != "Tourist" && role != "Guide" {
		http.Error(rw, "Access denied", http.StatusForbidden)
		return
	}
	fmt.Println("Getting followed users...")
	followed, err := fh.repo.GetFollowed(userID)
	if err != nil {
		fh.logger.Print("Database exception: ", err)
		http.Error(rw, "Database error", http.StatusInternalServerError)
		return
	}
	if len(followed) == 0 {
		rw.WriteHeader(http.StatusOK)
		emptyList := []data.User{}
		json.NewEncoder(rw).Encode(emptyList)
		return
	}
	err = followed.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		fh.logger.Print("Unable to convert to json :", err)
		return
	}
}

func (fh *FollowerHandler) Unfollow(rw http.ResponseWriter, h *http.Request) {
	userID, username, role, err := fh.getAuthorization(rw, h)
	fmt.Printf("Unfollowing for user ID: %d, Username: %s, Role: %s\n", userID, username, role)
	if err != nil {
		return
	}
	if role != "Tourist" && role != "Guide" {
		http.Error(rw, "Access denied", http.StatusForbidden)
		return
	}
	var user data.User
	err = user.FromJSON(h.Body)
	if err != nil {
		http.Error(rw, "Unable to parse json", http.StatusBadRequest)
		fh.logger.Print("Unable to parse json: ", err)
		return
	}

	unfollowedUser, err := fh.repo.Unfollow(&data.User{ID: userID, Username: username}, &user)
	if err != nil {
		if err.Error() == "no follow relationship found" {
			http.Error(rw, "No follow relationship found", http.StatusNotFound)
		} else {
			http.Error(rw, "Unable to unfollow user", http.StatusInternalServerError)
		}
		fh.logger.Print("Unable to unfollow user: ", err)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]string{
		"message": fmt.Sprintf("Unfollowed user: %s", unfollowedUser),
	})
}

func (fh *FollowerHandler) Follow(rw http.ResponseWriter, h *http.Request) {
	userID, username, role, err := fh.getAuthorization(rw, h)
	fmt.Printf("Following for user ID: %d, Username: %s, Role: %s\n", userID, username, role)
	if err != nil {
		return
	}
	if role != "Tourist" && role != "Guide" {
		http.Error(rw, "Access denied", http.StatusForbidden)
		return
	}
	var user data.User
	err = user.FromJSON(h.Body)
	if err != nil {
		http.Error(rw, "Unable to parse json", http.StatusBadRequest)
		fh.logger.Print("Unable to parse json: ", err)
		return
	}

	var followedUser *data.User
	followedUser, err = fh.repo.Follow(&data.User{ID: userID, Username: username}, &user)
	if err != nil {
		http.Error(rw, "Unable to create following", http.StatusInternalServerError)
		fh.logger.Print("Unable to create following: ", err)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]string{
		"message": fmt.Sprintf("Following user: %s", followedUser.Username),
	})
}

func (fh *FollowerHandler) GetSuggested(rw http.ResponseWriter, h *http.Request) {
	userID, username, role, err := fh.getAuthorization(rw, h)
	fmt.Printf("Getting suggested followers for user ID: %d, Username: %s, Role: %s\n", userID, username, role)
	if err != nil {
		return
	}
	if role != "Tourist" && role != "Guide" {
		http.Error(rw, "Access denied", http.StatusForbidden)
		return
	}
	fmt.Println("Getting suggested followers...")
	vars := mux.Vars(h)
	limitStr := vars["limit"]
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		fh.logger.Print("Invalid limit parameter: ", err)
		http.Error(rw, "Invalid limit parameter", http.StatusBadRequest)
		return
	}
	suggested, err := fh.repo.GetSuggested(userID, limit)
	if err != nil {
		fh.logger.Print("Database exception: ", err)
		http.Error(rw, "Database error", http.StatusInternalServerError)
		return
	}
	if len(suggested) == 0 {
		rw.WriteHeader(http.StatusOK)
		emptyList := []data.User{}
		json.NewEncoder(rw).Encode(emptyList)
		return
	}
	err = suggested.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		fh.logger.Print("Unable to convert to json :", err)
		return
	}
}
