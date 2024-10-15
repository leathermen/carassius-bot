package reddit

import (
	"errors"
	"log"

	"github.com/gocolly/colly"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/internal/helpers"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
)

var errFailedToGetImage = errors.New("failed to get reddit image")

func (rh *Handler) image(userID int64, url, redditID string, botname *telego.BotName) error {
	c := colly.NewCollector()

	var (
		success          bool
		errmsg, imageSRC string
	)

	c.OnHTML("body", func(e *colly.HTMLElement) {
		postImage := e.DOM.Find("#post-image")
		var hasSRC bool
		imageSRC, hasSRC = postImage.Attr("src")
		if !hasSRC {
			success = false
			errmsg = markuperrtxt
			return
		}

		success = true
	})

	c.OnError(func(_ *colly.Response, err error) {
		success = false
		log.Printf("reddit scraping faield: %s", err)
		errmsg = "Failed to get Reddit video :C"
	})

	if err := c.Visit(url); err != nil {
		log.Printf("failed to scrape reddit page: %s", err)
		return errFailedToGetVideo
	}

	if !success {
		return errors.New(errmsg)
	}

	file, err := helpers.DownloadFile(imageSRC, "jpg")
	if err != nil {
		log.Printf("failed to download reddit image: %s", err)
		return errFailedToGetImage
	}

	defer file.Close()

	tgMsg, err := rh.bot.SendPhoto(&telego.SendPhotoParams{
		ChatID: telegoutil.ID(userID),
		Photo:  telegoutil.File(file),
	})
	if err != nil {
		log.Printf("failed to send photo to user %d: %s", userID, err)
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Failed to download Reddit media!",
		}); err != nil {
			log.Printf("failed to send failed to download reddit media: %s", err)
		}
	}

	fileID := "nothing"

	for _, attachment := range tgMsg.Photo {
		fileID = attachment.FileID
		break
	}

	if fileID == "nothing" {
		log.Printf("failed to find TG file id")
		return errFailedToGetImage
	}

	for _, c := range rh.channels {
		if _, err = rh.bot.SendPhoto(&telego.SendPhotoParams{
			ChatID: telegoutil.ID(c),
			Photo:  telegoutil.FileFromID(fileID),
		}); err != nil {
			log.Printf("failed to send tg photo to a channel %d: %s", c, err)
		}
	}

	mediaFileData := queue.MediaFile{
		SocialNetworkID:   redditID,
		SocialNetworkName: Code,
		FileID:            fileID,
		FileType:          "image",
		Bot:               botname.Name,
	}

	if err := rh.db.InsertMediaFile(mediaFileData); err != nil {
		log.Printf("failed to save yt media post download: %s", err)
	}

	return nil
}
