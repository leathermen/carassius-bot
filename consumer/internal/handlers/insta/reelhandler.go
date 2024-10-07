package insta

import (
	"encoding/json"
	"log"
	"regexp"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/pkg/db"
	"github.com/nikitades/carassius-bot/shared/request"
)

type reelhandler struct {
	bot *telego.Bot
	db  db.Database
}

func newReelHandler(bot *telego.Bot, db db.Database) *reelhandler {
	return &reelhandler{bot, db}
}

func (rh *reelhandler) Handle(userID int64, msg string, msgID int) {
	reelID, found := getReelID(msg)

	if !found {
		log.Printf("failed to find reel ID")
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Malformed Instagram Reel link!",
		}); err != nil {
			log.Printf("failed to send malformed reel link message, user %d", userID)
		}
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

	req := &InstagramRequest{
		QueryHash: queryhash,
		ContentID: reelID,
		IsStory:   false,
	}

	cookie, found := rh.db.GetUserCookie()
	if !found {
		log.Printf("failed to get cookie!")
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Failed to download reel video",
		}); err != nil {
			log.Printf("failed to send failed to download reel video message, user %d", userID)
		}

		return
	}

	responseBody, err := makeInstagramRequest(req, cookie)
	if err != nil {
		log.Printf("failed to make ig reel request! user %d, url %s", userID, msg)
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Failed to download reel video",
		}); err != nil {
			log.Printf("failed to send failed to download reel video message, user %d", userID)
		}

		return
	}

	var InstagramResponseReels *InstagramResponseReels
	err = json.Unmarshal([]byte(responseBody), &InstagramResponseReels)
	if err != nil {
		log.Printf("failed to decore ig reel json")
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Failed to download reel video",
		}); err != nil {
			log.Printf("failed to send failed to download reel video message, user %d", userID)
		}

		return
	}

	//TODO роутинг GraphVideo, GraphImage
	log.Println("sudo done")
}

func getReelID(url string) (string, bool) {
	pattern := regexp.MustCompile(`(?i)/reel/([A-Za-z0-9_-]+)`)
	if match := pattern.FindStringSubmatch(url); len(match) > 1 {
		return match[1], true
	}

	return "", false
}
