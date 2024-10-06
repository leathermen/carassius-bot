package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikitades/carassius-bot/producer/internal/db"
	"github.com/nikitades/carassius-bot/producer/internal/router"
	"github.com/nikitades/carassius-bot/producer/pkg"
)

func main() {
	config := getConfig()

	dbURL := os.Getenv("DATABASE_URL")
	dbconn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("failed to create db connection")
	}

	db := db.New(dbconn)
	mediaRouter := router.New()

	bot := pkg.NewBot(config, mediaRouter, db, db)

	var signalChan chan (os.Signal) = make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-signalChan
		cancel()
	}()

	bot.Start(ctx)
}
