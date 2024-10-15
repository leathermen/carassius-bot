package reddit

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"

	"github.com/gocolly/colly"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/pkg/db"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
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

func (rh *Handler) Handle(userID int64, msg string, msgID int) {
	defer func() {
		if err := rh.q.DeleteMessageFromQueue(msgID); err != nil {
			log.Printf("failed to remove message from queue: %d", msgID)
		}
	}()

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

		shreddata := &ShredditData{}

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
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Failed to get Reddit page :C",
		}); err != nil {
			log.Printf("failed to send failed to get reddig page msg, user %d", userID)
		}

		return
	}

	if !success {
		log.Printf("failed to serve reddit: %s, post: %s", errmsg, msg)
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   errmsg,
		}); err != nil {
			log.Printf("failed to send <%s> msg, user %d", errmsg, userID)
		}

		return
	}

	switch typ {
	case "video":
		if err := rh.video(userID, msg); err != nil {
			if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
				ChatID: telegoutil.ID(userID),
				Text:   "Failed to download Reddit video :C",
			}); err != nil {
				log.Printf("failed to send failed to download reddit video, user %d", userID)
			}
		}
	case "image":
		if err := rh.image(userID, msg); err != nil {
			if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
				ChatID: telegoutil.ID(userID),
				Text:   "Failed to download Reddit image :C",
			}); err != nil {
				log.Printf("failed to send failed to download reddit image, user %d", userID)
			}
		}
	default:
		log.Printf("unsupported media type provided: %s, user %d", typ, userID)
		if _, err := rh.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "Unsupported Reddit media type!",
		}); err != nil {
			log.Printf("failed to send unsupported media type msg, user %d", userID)
		}
	}
}

func (rh *Handler) Name() string {
	return Code
}

func (rh *Handler) image(userID int64, url string) error {
	return nil
}
