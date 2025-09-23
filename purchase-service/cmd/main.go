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

	saga "github.com/tamararankovic/microservices_demo/common/saga/messaging"
	natsmsg "github.com/tamararankovic/microservices_demo/common/saga/messaging/nats"
)

func main() {
	database, err := db.NewDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	cfg := loadConfig()

	orchestratorPublisher := mustPublisher(cfg, cfg.CommandSubject)
	orchestratorSubscriber := mustSubscriber(cfg, cfg.ReplySubject, "orchestrator")

	orchestrator, err := services.NewTourPurchaseOrchestrator(orchestratorPublisher, orchestratorSubscriber)

	cartRepo := repositories.NewShoppingCartRepository(database)
	itemRepo := repositories.NewOrderItemRepository(database)
	tokenRepo := repositories.NewTourPurchaseTokenRepository(database)

	cartService := services.NewCartService(cartRepo, itemRepo)
	checkoutService := services.NewCheckoutService(cartRepo, itemRepo, tokenRepo, cartService)
	authService := services.NewAuthService()

	handlerPublisher := mustPublisher(cfg, cfg.ReplySubject)
	handlerSubscriber := mustSubscriber(cfg, cfg.CommandSubject, "purchase_handler_group")

	_, err = handlers.NewCreatePurchaseCommandHandler(checkoutService, handlerPublisher, handlerSubscriber)

	cartHandler := handlers.NewCartHandler(cartService, authService)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService, authService, orchestrator)

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

func mustPublisher(cfg config, subject string) saga.Publisher {
	pub, err := natsmsg.NewNATSPublisher(cfg.NatsHost, cfg.NatsPort, cfg.NatsUser, cfg.NatsPass, subject)
	if err != nil {
		log.Fatalf("nats publisher for subject %s: %v", subject, err)
	}
	return pub
}

func mustSubscriber(cfg config, subject string, queueGroup string) saga.Subscriber {
	sub, err := natsmsg.NewNATSSubscriber(cfg.NatsHost, cfg.NatsPort, cfg.NatsUser, cfg.NatsPass, subject, queueGroup)
	if err != nil {
		log.Fatalf("nats subscriber for subject %s: %v", subject, err)
	}
	return sub
}
