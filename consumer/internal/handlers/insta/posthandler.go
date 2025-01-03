package insta

import (
	"log"
	"regexp"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/internal/helpers"
	"github.com/nikitades/carassius-bot/consumer/pkg/db"
	"github.com/nikitades/carassius-bot/consumer/pkg/handler"
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

	channels []int64
}

func newPostHandler(bot *telego.Bot, db db.Database, csrfprovider *csrfprovider, channels []int64) *posthandler {
	return &posthandler{bot, db, csrfprovider, channels}
}

func (ph *posthandler) Handle(userID int64, msg string, _ int) error {
	postID, found := getPostID(msg)

	if !found {
		return handler.ErrFailedToGetMedia
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

		return nil
	}

	postDetails, err := getMediaDetails(postID, ph.csrfprovider.getCSRF())
	if err != nil {
		log.Printf("failed to get post details, post ID %s, user ID %d", postID, userID)
		return handler.ErrFailedToGetMedia
	}

	if len(postDetails.Data.Media.DisplayResources) == 0 {
		log.Printf("failed to find appropriate ig post image sizes")
		return handler.ErrFailedToGetMedia
	}

	image := postDetails.Data.Media.DisplayResources[0]
	for _, resource := range postDetails.Data.Media.DisplayResources {
		if resource.Height*resource.Width > image.Height*image.Width {
			image = resource
		}
	}

	file, err := helpers.DownloadFile(image.SRC)
	if err != nil {
		log.Printf("failed to download post image")
		return handler.ErrFailedToGetMedia
	}

	defer file.Close()

	inputfile := telego.InputFile{File: file}

	params := telego.SendPhotoParams{
		ChatID: telegoutil.ID(userID),
		Photo:  inputfile,
	}

	tgMsg, err := ph.bot.SendPhoto(&params)
	if err != nil {
		log.Printf("failed to send tg photo to user: %s", err)
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

	for _, c := range ph.channels {
		if _, err = ph.bot.SendPhoto(&telego.SendPhotoParams{
			ChatID: telegoutil.ID(c),
			Photo:  telegoutil.FileFromID(fileID),
		}); err != nil {
			log.Printf("failed to send tg photo to a channel %d: %s", c, err)
		}
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

	return nil
}
