package reddit

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"

	"github.com/gocolly/colly"
	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/consumer/pkg/db"
	"github.com/nikitades/carassius-bot/consumer/pkg/handler"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
	"github.com/nikitades/carassius-bot/shared/request"
)

const (
	Code         = "reddit"
	markuperrtxt = "Failed to understand Reddit markup :C"
)

type Handler struct {
	bot *telego.Bot
	q   queue.Queue
	db  db.Database

	channels []int64
}

func New(bot *telego.Bot, q queue.Queue, db db.Database, channels []int64) *Handler {
	return &Handler{bot, q, db, channels}
}

func (rh *Handler) Handle(userID int64, msg string, msgID int) error {
	defer func() {
		if err := rh.q.DeleteMessageFromQueue(msgID); err != nil {
			log.Printf("failed to remove message from queue: %d", msgID)
		}
	}()
	botname, _ := rh.bot.GetMyName(nil)
	redditID, err := extractRedditID(msg)
	if err != nil {
		log.Printf("failed to get reddit ID")
		return handler.ErrFailedToGetMedia
	}
	mediaFile, err := rh.db.GetMediaFileBySocialNetworkID(redditID, request.TypeReddit.String(), botname.Name)

	if err != nil {
		log.Printf("failed to lookup media files DB: %s", err)
	}

	if mediaFile != nil { //nolint:nestif
		if mediaFile.FileType == "video" {
			if _, err := rh.bot.SendVideo(&telego.SendVideoParams{
				ChatID: telego.ChatID{ID: userID},
				Video: telego.InputFile{
					FileID: mediaFile.FileID,
				},
			}); err != nil {
				log.Printf("failed to resend video to user %d: %s", userID, err)
			}

			for _, c := range rh.channels {
				if _, err := rh.bot.SendVideo(&telego.SendVideoParams{
					ChatID: telego.ChatID{ID: c},
					Video: telego.InputFile{
						FileID: mediaFile.FileID,
					},
				}); err != nil {
					log.Printf("failed to resend video to channel %d: %s", c, err)
				}
			}
		}

		if mediaFile.FileType == "image" {
			if _, err := rh.bot.SendPhoto(&telego.SendPhotoParams{
				ChatID: telego.ChatID{ID: userID},
				Photo: telego.InputFile{
					FileID: mediaFile.FileID,
				},
			}); err != nil {
				log.Printf("failed to resend video to user %d: %s", userID, err)
			}

			for _, c := range rh.channels {
				if _, err := rh.bot.SendPhoto(&telego.SendPhotoParams{
					ChatID: telego.ChatID{ID: c},
					Photo: telego.InputFile{
						FileID: mediaFile.FileID,
					},
				}); err != nil {
					log.Printf("failed to resend video to channel %d: %s", c, err)
				}
			}
		}

		return nil
	}

	var (
		success bool
		errmsg  string
		typ     string
	)

	c := colly.NewCollector()

	c.OnHTML("body", func(e *colly.HTMLElement) {
		shredditScreenview := e.DOM.Find("shreddit-screenview-data")
		dataRaw, hasData := shredditScreenview.Attr("data")
		if !hasData {
			success = false
			errmsg = markuperrtxt
			return
		}

		dataStr, _ := url.QueryUnescape(dataRaw)
		dataStr = strings.Replace(dataStr, "&quot;", "\"", -1)

		shreddata := &ShredditDataType{}

		if err := json.Unmarshal([]byte(dataStr), shreddata); err != nil {
			errmsg = markuperrtxt
			success = false
		}

		success = true
		typ = shreddata.Post.Type
	})

	// Handle errors during scraping
	c.OnError(func(_ *colly.Response, err error) {
		success = false
		log.Printf("reddit scraping failed: %s", err)
		errmsg = "Failed to get Reddit page :C"
	})

	if err := c.Visit(msg); err != nil {
		log.Printf("failed to scrape reddit page: %s", err)
		return handler.ErrFailedToGetMedia
	}

	if !success {
		log.Printf("failed to serve reddit: %s, post: %s", errmsg, msg)
		return handler.ErrFailedToGetMedia
	}

	switch typ {
	case "video":
		if err := rh.video(userID, msg, redditID, botname); err != nil {
			log.Printf("failed to download reddit video")
			return handler.ErrFailedToGetMedia
		}
	case "image":
		if err := rh.image(userID, msg, redditID, botname); err != nil {
			log.Printf("failed to download reddit image")
			return handler.ErrFailedToGetMedia
		}
	default:
		log.Printf("unsupported media type provided: %s, user %d", typ, userID)
		return handler.ErrUnsupported
	}

	return nil
}

func (rh *Handler) Name() string {
	return Code
}
