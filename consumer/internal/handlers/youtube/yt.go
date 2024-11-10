package youtube

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/kkdai/youtube/v2"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/pkg/db"
	"github.com/nikitades/carassius-bot/consumer/pkg/handler"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
	"github.com/nikitades/carassius-bot/shared/request"
	"github.com/thanhpk/randstr"
)

const (
	Code = "youtube"

	maxTGFileSize = 1024 * 1024 * 50
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

func (h *Handler) Name() string {
	return Code
}

func (h *Handler) Handle(userID int64, msg string, msgID int) error {
	defer func() {
		if err := h.q.DeleteMessageFromQueue(msgID); err != nil {
			log.Printf("failed to remove message from queue: %d", msgID)
		}
	}()

	videoID, ok := extractShortVideoID(msg)

	if !ok {
		log.Printf("failed to process youtube request: %s", msg)
		return handler.ErrFailedToGetMedia
	}

	botname, _ := h.bot.GetMyName(nil)
	mediaFile, err := h.db.GetMediaFileBySocialNetworkID(videoID, request.TypeYoutube.String(), botname.Name)

	if err != nil {
		log.Printf("failed to lookup media files DB: %s", err)
	}

	if mediaFile != nil {
		if _, err := h.bot.SendVideo(&telego.SendVideoParams{
			ChatID: telego.ChatID{ID: userID},
			Video: telego.InputFile{
				FileID: mediaFile.FileID,
			},
		}); err != nil {
			log.Printf("failed to resend video: %s", err)
		}

		return nil
	}

	client := youtube.Client{}

	video, err := client.GetVideo(msg)
	if err != nil {
		log.Printf("failed to download video: %s", err)
		return handler.ErrFailedToGetMedia
	}

	formats := video.Formats.WithAudioChannels()

	if len(formats) == 0 {
		log.Printf("no formats found with audio channels: %s", msg)
		return handler.ErrFailedToGetMedia
	}

	var optimalFormat youtube.Format

	for _, format := range formats {
		if strings.Contains(format.MimeType, "video/mp4") && format.ContentLength <= maxTGFileSize && format.ContentLength > optimalFormat.ContentLength {
			optimalFormat = format
		}
	}

	if optimalFormat.ContentLength > maxTGFileSize || optimalFormat.ContentLength == 0 {
		log.Printf("file size exceeded: %s", msg)
		return handler.ErrFailedToGetMedia
	}

	file, err := os.CreateTemp("", randstr.String(32)+".mp4")
	if err != nil {
		log.Printf("failed to create a tmp file: %v", err)
		return handler.ErrFailedToGetMedia
	}
	defer file.Close()

	// Download the video content
	stream, _, err := client.GetStream(video, &optimalFormat)
	if err != nil {
		log.Printf("failed to download video: %v", err)
		return handler.ErrFailedToGetMedia
	}

	_, err = io.Copy(file, stream)
	if err != nil {
		log.Printf("failed to copy video file: %v", err)
		return handler.ErrFailedToGetMedia
	}

	_, _ = file.Seek(0, 0) // ebal w ryt

	inputfile := telego.InputFile{File: file}

	params := telego.SendVideoParams{
		ChatID: telegoutil.ID(userID),
		Video:  inputfile,
		Width:  optimalFormat.Width,
		Height: optimalFormat.Height,
	}

	tgMsg, err := h.bot.SendVideo(&params)
	if err != nil {
		log.Printf("failed to send tg video to user: %s", err)
		return handler.ErrFailedToGetMedia
	}

	for _, c := range h.channels {
		if _, err = h.bot.SendVideo(&telego.SendVideoParams{
			ChatID: telegoutil.ID(c),
			Video:  telegoutil.FileFromID(tgMsg.Video.FileID),
			Width:  optimalFormat.Width,
			Height: optimalFormat.Height,
		}); err != nil {
			log.Printf("failed to send tg video to channel %d: %s", c, err)
		}
	}

	mediaFileData := queue.MediaFile{
		SocialNetworkID:   videoID,
		SocialNetworkName: Code,
		FileID:            tgMsg.Video.FileID,
		FileType:          "video",
		Bot:               botname.Name,
	}

	if err := h.db.InsertMediaFile(mediaFileData); err != nil {
		log.Printf("failed to save yt media post download: %s", err)
	}

	return nil
}
