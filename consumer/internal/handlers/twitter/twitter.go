package twitter

import (
	"log"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/internal/helpers"
	"github.com/nikitades/carassius-bot/consumer/pkg/db"
	"github.com/nikitades/carassius-bot/consumer/pkg/handler"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
	"github.com/nikitades/carassius-bot/shared/request"
)

const Code = "twitter"

type Handler struct {
	bot *telego.Bot
	db  db.Database
	q   queue.Queue

	channels []int64
}

func New(bot *telego.Bot, q queue.Queue, db db.Database, channels []int64) *Handler {
	return &Handler{bot, db, q, channels}
}

func (h *Handler) Handle(userID int64, msg string, msgID int) error { //nolint:gocyclo
	defer func() {
		if err := h.q.DeleteMessageFromQueue(msgID); err != nil {
			log.Printf("failed to remove message from queue: %d", msgID)
		}
	}()

	tweetID, ok := extractTweetID(msg)

	if !ok {
		log.Printf("failed to process twitter request: %s", msg)
		return handler.ErrFailedToGetMedia
	}

	botname, _ := h.bot.GetMyName(nil)
	mediaFile, err := h.db.GetMediaFileBySocialNetworkID(tweetID, request.TypeTwitter.String(), botname.Name)

	if err != nil {
		log.Printf("failed to lookup media files DB: %s", err)
	}

	if mediaFile != nil {
		switch mediaFile.FileType {
		case "video":
			if _, err := h.bot.SendVideo(&telego.SendVideoParams{
				ChatID: telego.ChatID{ID: userID},
				Video: telego.InputFile{
					FileID: mediaFile.FileID,
				},
			}); err != nil {
				log.Printf("failed to resend video: %s", err)
			}

		case "photo":
			if _, err := h.bot.SendPhoto(&telego.SendPhotoParams{
				ChatID: telego.ChatID{ID: userID},
				Photo: telego.InputFile{
					FileID: mediaFile.FileID,
				},
			}); err != nil {
				log.Printf("failed to resend photo: %s", err)
			}

		case "gif":
			if _, err := h.bot.SendAnimation(&telego.SendAnimationParams{
				ChatID: telego.ChatID{ID: userID},
				Animation: telego.InputFile{
					FileID: mediaFile.FileID,
				},
			}); err != nil {
				log.Printf("failed to resend gif: %s", err)
			}
		}

		return nil
	}

	guestToken, err := getGuestToken()
	if err != nil {
		log.Printf("failed to get twitter guest token: %s", err)
		return handler.ErrFailedToGetMedia
	}

	tweetDetails, err := getTweetDetails(tweetID, guestToken)
	if err != nil {
		log.Printf("failed to get tweet details: %s", err)
		return handler.ErrFailedToGetMedia
	}

	switch tweetDetails.typ {
	case twitterMediaTypeVideo:
		file, err := helpers.DownloadFile(tweetDetails.url)
		if err != nil {
			log.Printf("failed to download twitter attachment")
			return handler.ErrFailedToGetMedia
		}
		defer file.Close()

		inputfile := telego.InputFile{File: file}

		tgMsg, err := h.bot.SendVideo(&telego.SendVideoParams{
			ChatID: telegoutil.ID(userID),
			Video:  inputfile,
			Width:  tweetDetails.width,
			Height: tweetDetails.height,
		})
		if err != nil {
			log.Printf("failed to send twitter video, url %s, user %d", tweetDetails.url, userID)
			return handler.ErrFailedToGetMedia
		}

		for _, c := range h.channels {
			if _, err = h.bot.SendVideo(&telego.SendVideoParams{
				ChatID: telegoutil.ID(c),
				Video:  telegoutil.FileFromID(tgMsg.Video.FileID),
				Width:  tweetDetails.width,
				Height: tweetDetails.height,
			}); err != nil {
				log.Printf("failed to send tg video to a channel %d: %s", c, err)
			}
		}

		mediaFileData := queue.MediaFile{
			SocialNetworkID:   tweetID,
			SocialNetworkName: Code,
			FileID:            tgMsg.Video.FileID,
			FileType:          "video",
			Bot:               botname.Name,
		}

		if err := h.db.InsertMediaFile(mediaFileData); err != nil {
			log.Printf("failed to save twitter media post download: %s", err)
		}
	case twitterMediaTypePhoto:
		file, err := helpers.DownloadFile(tweetDetails.url)
		if err != nil {
			log.Printf("failed to download twitter attachment")
			return handler.ErrFailedToGetMedia
		}
		defer file.Close()

		inputfile := telego.InputFile{File: file}

		tgMsg, err := h.bot.SendPhoto(&telego.SendPhotoParams{
			ChatID: telegoutil.ID(userID),
			Photo:  inputfile,
		})
		if err != nil {
			log.Printf("failed to send twitter photo, url %s, user %d", tweetDetails.url, userID)
			return handler.ErrFailedToGetMedia
		}

		fileID := "nothing"

		for _, attachment := range tgMsg.Photo {
			fileID = attachment.FileID
			break
		}

		if fileID == "nothing" {
			log.Printf("failed to find TG file id")
			return handler.ErrFailedToGetMedia
		}

		for _, c := range h.channels {
			if _, err = h.bot.SendPhoto(&telego.SendPhotoParams{
				ChatID: telegoutil.ID(c),
				Photo:  telegoutil.FileFromID(fileID),
			}); err != nil {
				log.Printf("failed to send tg photo to a channel %d: %s", c, err)
			}
		}

		mediaFileData := queue.MediaFile{
			SocialNetworkID:   tweetID,
			SocialNetworkName: Code,
			FileID:            fileID,
			FileType:          "photo",
			Bot:               botname.Name,
		}

		if err := h.db.InsertMediaFile(mediaFileData); err != nil {
			log.Printf("failed to save twitter media post download: %s", err)
		}
	case twitterMediaTypeGif:
		file, err := helpers.DownloadFile(tweetDetails.url, tweetDetails.url[len(tweetDetails.url)-3:])
		if err != nil {
			log.Printf("failed to download twitter attachment")
			return handler.ErrFailedToGetMedia
		}
		defer file.Close()

		inputfile := telego.InputFile{File: file}

		tgMsg, err := h.bot.SendAnimation(&telego.SendAnimationParams{
			ChatID:    telegoutil.ID(userID),
			Animation: inputfile,
			Width:     tweetDetails.width,
			Height:    tweetDetails.height,
		})
		if err != nil {
			log.Printf("failed to send twitter gif, url %s, user %d", tweetDetails.url, userID)
			return handler.ErrFailedToGetMedia
		}

		for _, c := range h.channels {
			if _, err = h.bot.SendAnimation(&telego.SendAnimationParams{
				ChatID:    telegoutil.ID(c),
				Animation: telegoutil.FileFromID(tgMsg.Document.FileID),
				Width:     tweetDetails.width,
				Height:    tweetDetails.height,
			}); err != nil {
				log.Printf("failed to send tg gif to a channel %d: %s", c, err)
			}
		}

		mediaFileData := queue.MediaFile{
			SocialNetworkID:   tweetID,
			SocialNetworkName: Code,
			FileID:            tgMsg.Document.FileID,
			FileType:          "gif",
			Bot:               botname.Name,
		}

		if err := h.db.InsertMediaFile(mediaFileData); err != nil {
			log.Printf("failed to save twitter media post download: %s", err)
		}
	}
	return nil
}

func (h *Handler) Name() string {
	return Code
}
