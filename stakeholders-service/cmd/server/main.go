package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"stakeholders-service/db"
	"stakeholders-service/internal/handlers"
	repository "stakeholders-service/internal/repositories"
	proto "stakeholders-service/proto"
	"stakeholders-service/tracing"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

const serviceName = "stakeholders-service"

var tp *sdktrace.TracerProvider

func main() {
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	var shutdown func(context.Context) error
	tp, shutdown, err = tracing.Init(serviceName)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	defer database.Close()

	userRepo := repository.NewUserRepository(database)
	userHandler := handlers.NewUserHandler(userRepo)

	positionRepo := repository.NewPositionRepository(database)
	positionHandler := handlers.NewPositionHandler(positionRepo, userRepo)

	router := mux.NewRouter().StrictSlash(true)
	router.Use(otelmux.Middleware(serviceName))

	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}).Methods("GET")
	router.HandleFunc("/weapons", weaponGetHandler).Methods("GET")
	router.HandleFunc("/api/register", userHandler.RegisterUser).Methods("POST")
	router.HandleFunc("/api/login", userHandler.LoginUser).Methods("POST")
	router.HandleFunc("/api/profile", userHandler.GetMyProfile).Methods("GET")
	router.HandleFunc("/api/validateRole", userHandler.ValidateRole).Methods("POST")
	router.HandleFunc("/api/updateProfile", userHandler.UpdateMyProfile).Methods("PUT")
	router.HandleFunc("/api/me", userHandler.GetUserFromToken).Methods("GET")

	router.HandleFunc("/api/profile/{id}", userHandler.GetUserProfile).Methods("GET")
	router.HandleFunc("/api/usersForSearch", userHandler.GetUsersForSearch).Methods("GET")

	router.HandleFunc("/api/admin/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/api/admin/users/block", userHandler.BlockUser).Methods("PUT")

	router.HandleFunc("/api/position", positionHandler.GetPosition).Methods("GET")
	router.HandleFunc("/api/position", positionHandler.CreatePosition).Methods("POST")

	fmt.Println("Registered HTTP routes:")
	_ = router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		t, _ := route.GetPathTemplate()
		m, _ := route.GetMethods()
		fmt.Printf(" - %s %v\n", t, m)
		return nil
	})
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

func weaponGetHandler(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer(serviceName).Start(r.Context(), "weapon-get")
	defer span.End()

	span.AddEvent("Establishing connection to the database")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))

	span.SetStatus(codes.Ok, "ok")
}

func httpError(err error, span trace.Span, w http.ResponseWriter, status int) {
	log.Println(err.Error())
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	// Koristimo http.Error za slanje greške
	http.Error(w, err.Error(), status)
}

func httpErrorInternalServerError(err error, span trace.Span, w http.ResponseWriter) {
	httpError(err, span, w, http.StatusInternalServerError)
}
