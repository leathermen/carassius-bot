package pkg

import (
	"errors"
	"fmt"
	"log"

	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/producer/pkg/router"
	"github.com/nikitades/carassius-bot/shared/bothelper"
	"github.com/nikitades/carassius-bot/shared/request"
)

func (b *Producer) Handle(update telego.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	if update.Message.Text == "/start" {
		b.start(update)
		return
	}

	mtype, err := b.router.Route(update.Message.Text)

	if err != nil {
		if errors.Is(err, router.ErrNoMedia) {
			if _, err = b.bot.SendMessage(&telego.SendMessageParams{
				ChatID: update.Message.Chat.ChatID(),
				Text:   NoLinkOrUnknownMedia,
			}); err != nil {
				log.Printf("failed to send unknown media msg, user %d, bot %s: %s", update.Message.From.ID, bothelper.Botname(b.bot), err)
			}
		}
		return
	}

	exists, err := b.db.UserExistsInDB(update.Message.From.ID)
	if err != nil {
		log.Printf("failed to check the user for existence: %d: %s", update.Message.From.ID, err)
	}

	if !exists {
		b.registerUser(update)
	} else {
		b.updateUser(update)
	}

	if err = b.db.AddUserMessageToDB(
		update.Message.From.ID,
		update.Message.From.FirstName,
		update.Message.From.LastName,
		update.Message.From.Username,
		update.Message.From.LanguageCode,
		update.Message.Text,
	); err != nil {
		log.Printf("failed to add user message to db, user %d, bot %s: %s", update.Message.From.ID, bothelper.Botname(b.bot), err)
	}

	if mtype == request.TypeThanks {
		b.thanks(update)
		return
	}

	if err = b.publisher.AddMessageToQueue(update.Message.From.ID, update.Message.Text, bothelper.Botname(b.bot), mtype.String()); err != nil {
		log.Printf("failed to add msg to queue, user %d, bot %s: %s", update.Message.From.ID, bothelper.Botname(b.bot), err)
	}

	count, err := b.publisher.GetMessagesCountByBot(bothelper.Botname(b.bot))

	if err != nil {
		log.Printf("failed to get remaining messages count: %s", err)
		return
	}

	if _, err = b.bot.SendMessage(&telego.SendMessageParams{
		ChatID: update.Message.Chat.ChatID(),
		Text:   fmt.Sprintf("Don't block the bot. Wait. Your request is in the queue. There are currently %d messages ahead of yours.", count),
	}); err != nil {
		log.Printf("failed to send success msg, user %d, bot %s: %s", update.Message.From.ID, bothelper.Botname(b.bot), err)
	}
}
