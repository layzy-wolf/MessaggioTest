package main

import (
	"MessagioTest/config"
	"MessagioTest/internal/service/receiver"
	http2 "MessagioTest/internal/transport/http"
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	db, err := receiver.SetupDB(&cfg)
	if err != nil {
		log.Fatalf("error while creating db: %v", err)
	}

	producer, err := receiver.SetupProducer(&cfg)
	if err != nil {
		log.Fatalf("error while connecting to producer: %v", err)
	}

	messages, c := receiver.SetupReceiver()
	defer c()

	go receiver.SetupSender(ctx, messages, producer, &cfg)

	r := http2.SetupApi(db, messages)

	srv := &http.Server{
		Addr:    cfg.Receiver.Socket,
		Handler: r,
	}

	go func() {
		log.Printf("server started on port %v", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("Server shutting down gracefully, press Ctrl+C to force")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
