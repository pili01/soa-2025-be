package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gateway/internal/config"
	"gateway/internal/router"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// LOGOVI 
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Starting SOA Gateway...")

	//KONFIGURACIJAA
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	//RUTER
	r, err := router.NewRouter(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create router")
	}

	// SERVER
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r.GetEngine(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// STARTVOANJE SERVERA
	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("Gateway server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down gateway server...")


	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Gateway server exited")
}

