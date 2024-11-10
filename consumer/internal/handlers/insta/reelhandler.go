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

type reelhandler struct {
	bot          *telego.Bot
	db           db.Database
	csrfprovider *csrfprovider

	channels []int64
}

func newReelHandler(bot *telego.Bot, db db.Database, csrfprovider *csrfprovider, channels []int64) *reelhandler {
	return &reelhandler{bot, db, csrfprovider, channels}
}

func (rh *reelhandler) Handle(userID int64, msg string, _ int) error {
	reelID, found := getReelID(msg)

	if !found {
		return handler.ErrFailedToGetMedia
	}

	botname, _ := rh.bot.GetMyName(nil)
	mediaFile, err := rh.db.GetMediaFileBySocialNetworkID(reelID, request.TypeInsta.String(), botname.Name)

	if err != nil {
		log.Printf("failed to lookup media filed DB: %s", err)
	}

	if mediaFile != nil {
		if _, err := rh.bot.SendVideo(&telego.SendVideoParams{
			ChatID: telego.ChatID{ID: userID},
			Video: telego.InputFile{
				FileID: mediaFile.FileID,
			},
		}); err != nil {
			log.Printf("failed to resend video: %s", err)
		}

		return nil
	}

	reelDetails, err := getMediaDetails(reelID, rh.csrfprovider.getCSRF())
	if err != nil {
		return handler.ErrFailedToGetMedia
	}

	file, err := helpers.DownloadFile(reelDetails.Data.Media.VideoURL)
	if err != nil {
		log.Printf("failed to download reels video: %s", err)
		return handler.ErrFailedToGetMedia
	}

	defer file.Close()

	inputfile := telego.InputFile{File: file}

	params := telego.SendVideoParams{
		ChatID: telegoutil.ID(userID),
		Video:  inputfile,
		Width:  reelDetails.Data.Media.Dimensions.Width,
		Height: reelDetails.Data.Media.Dimensions.Height,
	}

	tgMsg, err := rh.bot.SendVideo(&params)
	if err != nil {
		log.Printf("failed to send tg video to user: %s", err)
		return handler.ErrFailedToGetMedia
	}

	for _, c := range rh.channels {
		if _, err = rh.bot.SendVideo(&telego.SendVideoParams{
			ChatID: telegoutil.ID(c),
			Video:  telegoutil.FileFromID(tgMsg.Video.FileID),
			Width:  reelDetails.Data.Media.Dimensions.Width,
			Height: reelDetails.Data.Media.Dimensions.Height,
		}); err != nil {
			log.Printf("failed to send tg video to a channel %d: %s", c, err)
		}
	}

	mediaFileData := queue.MediaFile{
		SocialNetworkID:   reelID,
		SocialNetworkName: Code,
		FileID:            tgMsg.Video.FileID,
		FileType:          "video",
		Bot:               botname.Name,
	}

	if err := rh.db.InsertMediaFile(mediaFileData); err != nil {
		log.Printf("failed to save yt media post download: %s", err)
	}

	return nil
}

func getReelID(url string) (string, bool) {
	pattern := regexp.MustCompile(`(?i)/reel/([A-Za-z0-9_-]+)`)
	if match := pattern.FindStringSubmatch(url); len(match) > 1 {
		return match[1], true
	}

	return "", false
}
