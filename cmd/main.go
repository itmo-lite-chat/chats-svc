package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/itmo-lite-chat/chats-svc/cmd/config"
	"github.com/itmo-lite-chat/chats-svc/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New(config.New())
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
