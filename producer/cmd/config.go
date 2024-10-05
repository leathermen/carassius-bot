package main

import (
	"os"

	"github.com/nikitades/carassius-bot/producer/pkg"
)

func getConfig() *pkg.BotConfig {
	token := os.Getenv("TOKEN")

	return &pkg.BotConfig{
		Token: token,
	}
}
