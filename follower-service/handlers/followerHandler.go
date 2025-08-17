package handlers

import (
	"encoding/json"
	"fmt"
	"follower-service/data"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type FollowerHandler struct {
	logger *log.Logger
	repo   *data.FollowerRepo
}

func NewFollowerHandler(l *log.Logger, r *data.FollowerRepo) *FollowerHandler {
	return &FollowerHandler{
		logger: l,
		repo:   r,
	}
}

func (fh *FollowerHandler) GetFollowers(rw http.ResponseWriter, h *http.Request) {
	fmt.Println("Getting followers...")
	followers, err := fh.repo.GetFollowers(5)
	if err != nil {
		fh.logger.Print("Database exception: ", err)
		http.Error(rw, "Database error", http.StatusInternalServerError)
		return
	}
	if len(followers) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("No followers found"))
		return
	}
	err = followers.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		fh.logger.Print("Unable to convert to json :", err)
		return
	}
}

func (fh *FollowerHandler) IsFollowedByMe(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	followedIdStr := vars["id"]
	fmt.Printf("Checking if followed by me for user id: %s\n", followedIdStr)
	followedId, err := strconv.Atoi(followedIdStr)
	if err != nil {
		fh.logger.Print("Invalid id parameter: ", err)
		http.Error(rw, "Invalid id parameter", http.StatusBadRequest)
		return
	}
	followers, err := fh.repo.IsFollowedByMe(3, followedId)
	if err != nil {
		fh.logger.Print("Database exception: ", err)
		http.Error(rw, "Database error", http.StatusInternalServerError)
		return
	}
	if len(followers) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(`{"value": false}`))
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte(`{"value": true}`))
}

func (fh *FollowerHandler) GetFollowed(rw http.ResponseWriter, h *http.Request) {
	fmt.Println("Getting followed users...")
	followed, err := fh.repo.GetFollowed(3)
	if err != nil {
		fh.logger.Print("Database exception: ", err)
		http.Error(rw, "Database error", http.StatusInternalServerError)
		return
	}
	if len(followed) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("No followed users found"))
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
	var user data.User
	err := user.FromJSON(h.Body)
	if err != nil {
		http.Error(rw, "Unable to parse json", http.StatusBadRequest)
		fh.logger.Print("Unable to parse json: ", err)
		return
	}

	unfollowedUser, err := fh.repo.Unfollow(&user, &data.User{ID: 5, Username: "user1"})
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
	var user data.User
	err := user.FromJSON(h.Body)
	if err != nil {
		http.Error(rw, "Unable to parse json", http.StatusBadRequest)
		fh.logger.Print("Unable to parse json: ", err)
		return
	}

	var followedUser *data.User
	followedUser, err = fh.repo.Follow(&user, &data.User{ID: 5, Username: "user1"})
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
	fmt.Println("Getting suggested followers...")
	suggested, err := fh.repo.GetSuggested(3, 3)
	if err != nil {
		fh.logger.Print("Database exception: ", err)
		http.Error(rw, "Database error", http.StatusInternalServerError)
		return
	}
	if len(suggested) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("No suggested followers found"))
		return
	}
	err = suggested.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		fh.logger.Print("Unable to convert to json :", err)
		return
	}
}
