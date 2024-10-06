package bothelper

import (
	"log"

	"github.com/mymmrac/telego"
)

func Botname(bot *telego.Bot) string {
	botname, err := bot.GetMyName(nil)
	if err != nil {
		log.Fatalf("failed to get own bot's name at registering the user: %s", err)
	}

	return botname.Name
}
