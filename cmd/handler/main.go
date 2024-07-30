package main

import (
	"MessagioTest/config"
	"MessagioTest/internal/service/handler"
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	db, err := handler.SetupDB(&cfg)
	if err != nil {
		log.Fatalf("error while creating db: %v", err)
	}

	go handler.SetupConsumerGroup(ctx, &cfg, db)

	log.Println("server started")

	<-ctx.Done()
	stop()
	log.Println("Server shutting down gracefully, press Ctrl+C to force")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	log.Println("Server exiting")
}
