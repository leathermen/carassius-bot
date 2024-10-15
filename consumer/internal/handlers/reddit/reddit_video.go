package reddit

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/gocolly/colly"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/internal/helpers"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
)

var errFailedToGetVideo = errors.New("failed to get Reddit video")

func (rh *Handler) video(userID int64, url, redditID string, botname *telego.BotName) error {
	c := colly.NewCollector()

	var (
		success             bool
		errmsg, mediaPacked string
	)

	c.OnHTML("body", func(e *colly.HTMLElement) {
		shredditScreenview := e.DOM.Find("shreddit-player-2")
		var hasMedia bool
		mediaPacked, hasMedia = shredditScreenview.Attr("packaged-media-json")
		if !hasMedia {
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

	mediaJSON := &ShredditDataVideo{}
	err := json.Unmarshal([]byte(mediaPacked), mediaJSON)
	if err != nil {
		log.Printf("failed to unmarshall reddit shreddit: %s", err)
		return errFailedToGetVideo
	}

	if len(mediaJSON.PlaybackMP4S.Permutations) == 0 {
		log.Printf("no video variants found at all")
		return errFailedToGetVideo
	}

	currentSource := mediaJSON.PlaybackMP4S.Permutations[0]
	for _, source := range mediaJSON.PlaybackMP4S.Permutations {
		if (source.Source.Dimensions.Height * source.Source.Dimensions.Width) > (currentSource.Source.Dimensions.Height * currentSource.Source.Dimensions.Width) {
			currentSource = source
		}
	}

	file, err := helpers.DownloadFile(currentSource.Source.URL, currentSource.Source.URL[len(currentSource.Source.URL)-3:])
	if err != nil {
		log.Printf("failed to download reddit video: %s", err)
		return errFailedToGetVideo
	}
	defer file.Close()

	inputfile := telego.InputFile{File: file}

	params := telego.SendVideoParams{
		ChatID: telegoutil.ID(userID),
		Video:  inputfile,
		Width:  currentSource.Source.Dimensions.Width,
		Height: currentSource.Source.Dimensions.Height,
	}

	tgMsg, err := rh.bot.SendVideo(&params)
	if err != nil {
		log.Printf("failed to send tg video to user: %s", err)

		return errFailedToGetVideo
	}

	for _, c := range rh.channels {
		if _, err = rh.bot.SendVideo(&telego.SendVideoParams{
			ChatID: telegoutil.ID(c),
			Video:  telegoutil.FileFromID(tgMsg.Video.FileID),
			Width:  currentSource.Source.Dimensions.Width,
			Height: currentSource.Source.Dimensions.Height,
		}); err != nil {
			log.Printf("failed to send tg video to channel %d: %s", c, err)
		}
	}

	mediaFileData := queue.MediaFile{
		SocialNetworkID:   redditID,
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
