package insta

import (
	"log"
	"regexp"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/internal/helpers"
	"github.com/nikitades/carassius-bot/consumer/pkg/db"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
	"github.com/nikitades/carassius-bot/shared/request"
)

func getPostID(url string) (string, bool) {
	pattern := regexp.MustCompile(`(?i)/p/([A-Za-z0-9_-]+)`)
	if match := pattern.FindStringSubmatch(url); len(match) > 1 {
		return match[1], true
	}

	return "", false
}

type posthandler struct {
	bot          *telego.Bot
	db           db.Database
	csrfprovider *csrfprovider
}

func newPostHandler(bot *telego.Bot, db db.Database, csrfprovider *csrfprovider) *posthandler {
	return &posthandler{bot, db, csrfprovider}
}

func (ph *posthandler) Handle(userID int64, msg string, _ int) {
	postID, found := getPostID(msg)

	if !found {
		log.Printf("failed to find post ID")
		if _, err := ph.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Malformed Instagram post link!",
		}); err != nil {
			log.Printf("failed to send malformed post link message, user %d", userID)
		}
	}

	botname, _ := ph.bot.GetMyName(nil)
	mediaFile, err := ph.db.GetMediaFileBySocialNetworkID(postID, request.TypeInsta.String(), botname.Name)

	if err != nil {
		log.Printf("failed to lookup media filed DB: %s", err)
	}

	if mediaFile != nil {
		if _, err := ph.bot.SendPhoto(&telego.SendPhotoParams{
			ChatID: telego.ChatID{ID: userID},
			Photo: telego.InputFile{
				FileID: mediaFile.FileID,
			},
		}); err != nil {
			log.Printf("failed to resend video: %s", err)
		}

		return
	}

	postDetails, err := getMediaDetails(postID, ph.csrfprovider.getCSRF())
	if err != nil {
		log.Printf("failed to get post details, post ID %s, user ID %d", postID, userID)
		if _, err := ph.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Failed to download post",
		}); err != nil {
			log.Printf("failed to send failed to download photo, post ID %s, user ID %d", postID, userID)
		}
	}

	file, err := helpers.DownloadFile(postDetails.Data.Media.Thumbnail)
	if err != nil {
		log.Printf("failed to download post image")
		if _, err := ph.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Failed to download post",
		}); err != nil {
			log.Printf("failed to send failed to download photo, post ID %s, userID %d", postID, userID)
		}
	}

	inputfile := telego.InputFile{File: file}

	params := telego.SendPhotoParams{
		ChatID: telegoutil.ID(userID),
		Photo:  inputfile,
	}

	tgMsg, err := ph.bot.SendPhoto(&params)
	if err != nil {
		log.Printf("failed to send tg photo: %s", err)
		if _, err := ph.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telego.ChatID{ID: userID},
			Text:   "Can't download this media!",
		}); err != nil {
			log.Printf("failed to send can't download msg, user %d: %s", userID, err)
		}

		return
	}

	fileID := "nothing"

	for _, attachment := range tgMsg.Photo {
		fileID = attachment.FileID
		break
	}

	if fileID == "nothing" {
		log.Printf("failed to find TG file id")
		return
	}

	mediaFileData := queue.MediaFile{
		SocialNetworkID:   postID,
		SocialNetworkName: Code,
		FileID:            fileID,
		FileType:          "image",
		Bot:               botname.Name,
	}

	if err := ph.db.InsertMediaFile(mediaFileData); err != nil {
		log.Printf("failed to save yt media post download: %s", err)
	}
}
