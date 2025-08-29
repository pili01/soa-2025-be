package main

import (
	"log"
	"net/http"
	"os"

	"purchase-service/db"
	"purchase-service/internal/handlers"
	"purchase-service/internal/repositories"
	"purchase-service/internal/services"

	"github.com/gorilla/mux"
)

func main() {
	database, err := db.NewDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	cartRepo := repositories.NewShoppingCartRepository(database)
	itemRepo := repositories.NewOrderItemRepository(database)
	tokenRepo := repositories.NewTourPurchaseTokenRepository(database)

	cartService := services.NewCartService(cartRepo, itemRepo)
	checkoutService := services.NewCheckoutService(cartRepo, itemRepo, tokenRepo)
	authService := services.NewAuthService()

	cartHandler := handlers.NewCartHandler(cartService, authService)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService, authService)

	router := mux.NewRouter()

	router.HandleFunc("/cart/items", cartHandler.AddToCart).Methods("POST")
	router.HandleFunc("/cart/items", cartHandler.RemoveFromCart).Methods("DELETE")
	router.HandleFunc("/cart", cartHandler.GetCart).Methods("GET")

	router.HandleFunc("/checkout", checkoutHandler.ProcessCheckout).Methods("POST")
	router.HandleFunc("/purchases", checkoutHandler.GetPurchaseHistory).Methods("GET")
	router.HandleFunc("/validate-token", checkoutHandler.ValidateToken).Methods("GET")
	router.HandleFunc("/check-is-purchased/{tourId}", checkoutHandler.CheckIsPurchased).Methods("GET")

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting purchase service on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
