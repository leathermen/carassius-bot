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

type reelhandler struct {
	bot          *telego.Bot
	db           db.Database
	csrfprovider *csrfprovider

	channels []int64
}

func newReelHandler(bot *telego.Bot, db db.Database, csrfprovider *csrfprovider, channels []int64) *reelhandler {
	return &reelhandler{bot, db, csrfprovider, channels}
}

func (rh *reelhandler) Handle(userID int64, msg string, _ int) {
	reelID, found := getReelID(msg)

	if !found {
		log.Printf("failed to find reel ID")
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Malformed Instagram Reel link!",
		}); err != nil {
			log.Printf("failed to send malformed reel link message, user %d", userID)
		}

		return
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

		return
	}

	reelDetails, err := getMediaDetails(reelID, rh.csrfprovider.getCSRF())
	if err != nil {
		log.Printf("failed to get reels details, reel ID %s, user ID %d", reelID, userID)
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Failed to download reel",
		}); err != nil {
			log.Printf("failed to send failed to download video, reel ID %s, user ID %d", reelID, userID)
		}

		return
	}

	file, err := helpers.DownloadFile(reelDetails.Data.Media.VideoURL)
	if err != nil {
		log.Printf("failed to download reels video: %s", err)
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Failed to download reel",
		}); err != nil {
			log.Printf("failed to send failed to download video, reel ID %s, userID %d", reelID, userID)
		}

		return
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
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telego.ChatID{ID: userID},
			Text:   "Can't download this media!",
		}); err != nil {
			log.Printf("failed to send can't download msg, user %d: %s", userID, err)
		}

		return
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
}

func getReelID(url string) (string, bool) {
	pattern := regexp.MustCompile(`(?i)/reel/([A-Za-z0-9_-]+)`)
	if match := pattern.FindStringSubmatch(url); len(match) > 1 {
		return match[1], true
	}

	return "", false
}
