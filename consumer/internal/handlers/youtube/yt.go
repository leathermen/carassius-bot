package youtube

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
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
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"golang.org/x/net/proxy"
)

const (
	Code = "youtube"

	maxVideoFileSize  = 1024 * 1024 * 2
	maxAudioTrackSize = 1024 * 1024 * 1.5

	maxJointFileSize = 1024 * 1024 * 5
)

type Handler struct {
	bot     *telego.Bot
	q       queue.Queue
	db      db.Database
	pp      *handler.Proxy
	botname *telego.BotName

	channels []int64
}

func New(bot *telego.Bot, q queue.Queue, db db.Database, proxyParams *handler.Proxy, channels []int64) *Handler {
	handler := &Handler{
		bot:      bot,
		q:        q,
		db:       db,
		pp:       proxyParams,
		channels: channels,
		botname:  nil,
	}

	botname, _ := handler.bot.GetMyName(nil)
	handler.botname = botname

	return handler
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

	mediaFile, err := h.db.GetMediaFileBySocialNetworkID(videoID, request.TypeYoutube.String(), h.botname.Name)

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

	var socks5dialer proxy.Dialer

	for range 5 {
		socks5dialer, err = proxy.SOCKS5("tcp", h.pp.HostnamePort(), &proxy.Auth{
			User:     h.pp.UsernameWithCountryAndRandomSID(),
			Password: h.pp.Password,
		}, proxy.Direct)

		if err == nil {
			break
		}
	}

	if err != nil {
		log.Printf("failed to connect to proxy: %s", err)
		return handler.ErrFailedToGetMedia
	}

	transport := http.Transport{
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			conn, err := socks5dialer.Dial(network, addr)
			if err != nil {
				return conn, fmt.Errorf("failed to dial with proxy: %w", err)
			}

			return conn, nil
		},
	}

	client := youtube.Client{
		HTTPClient: &http.Client{
			Transport: &transport,
		},
	}

	video, err := client.GetVideo(msg)
	if err != nil {
		log.Printf("failed to download video: %s", err)
		return handler.ErrFailedToGetMedia
	}

	separateVideo, separateAudio, found := h.findSeparateFormats(video)

	if found {
		return h.replyMerged(video, &separateVideo, &separateAudio, userID, &client)
	}

	singleVideo, found := h.findSingleFormat(video)

	if found {
		return h.replySingle(video, &singleVideo, userID, &client)
	}

	log.Printf("failed to find appropriate youtube video formats")
	return handler.ErrFailedToGetMedia
}

func (h *Handler) findSeparateFormats(video *youtube.Video) (videoFormat, audioFormat youtube.Format, found bool) {
	var optimalVideoFormat youtube.Format
	var optimalAudioFormat youtube.Format

	for _, v := range video.Formats {
		isMP4 := strings.Contains(v.MimeType, "video/mp4")
		isLessThanLimit := v.ContentLength <= maxVideoFileSize
		isMoreThanTheLastOne := v.ContentLength > optimalVideoFormat.ContentLength
		if isMP4 && isLessThanLimit && isMoreThanTheLastOne {
			optimalVideoFormat = v
		}
	}

	for _, a := range video.Formats {
		isAudioMP4 := strings.Contains(a.MimeType, "audio/mp4")
		isLessThanLimit := a.ContentLength <= maxAudioTrackSize
		isMoreThanTheLastOne := a.ContentLength > optimalAudioFormat.ContentLength
		if isLessThanLimit && isMoreThanTheLastOne && isAudioMP4 {
			optimalAudioFormat = a
		}
	}

	if optimalVideoFormat.ContentLength == 0 {
		log.Printf("no appropriate separate video track found")
		return youtube.Format{}, youtube.Format{}, false
	}

	if optimalAudioFormat.ContentLength == 0 {
		log.Printf("no appropriate separate audio track found")
		return optimalVideoFormat, youtube.Format{}, false
	}

	return optimalVideoFormat, optimalAudioFormat, true
}

func (h *Handler) findSingleFormat(video *youtube.Video) (youtube.Format, bool) {
	var optimalVideoFormat youtube.Format

	withAudio := []youtube.Format{}

	for _, v := range video.Formats {
		isMP4 := strings.Contains(v.MimeType, "video/mp4")
		isLessThanLimit := v.ContentLength <= maxJointFileSize
		isMoreThanTheLastOne := v.ContentLength > optimalVideoFormat.ContentLength
		hasAudioTrack := v.AudioQuality != ""
		if hasAudioTrack {
			withAudio = append(withAudio, v)
		}
		if isMP4 && isLessThanLimit && isMoreThanTheLastOne && hasAudioTrack {
			optimalVideoFormat = v
		}
	}

	if optimalVideoFormat.ContentLength == 0 {
		log.Printf("no appropriate joint video track found")
		return youtube.Format{}, false
	}

	return optimalVideoFormat, true
}

func (h *Handler) replyMerged(video *youtube.Video, videoFormat, audioFormat *youtube.Format, userID int64, client *youtube.Client) error {
	file, err := os.CreateTemp("", randstr.String(32)+".mp4")
	if err != nil {
		log.Printf("failed to create a tmp video file: %v", err)
		return handler.ErrFailedToGetMedia
	}
	defer file.Close()

	// Download the video content
	stream, _, err := client.GetStream(video, videoFormat)
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

	audiofile, err := os.CreateTemp("", randstr.String(32)+".m4a")
	if err != nil {
		log.Printf("failed to create a tmp audio file: %v", err)
		return handler.ErrFailedToGetMedia
	}
	defer audiofile.Close()

	// Download the audio content
	stream, _, err = client.GetStream(video, audioFormat)
	if err != nil {
		log.Printf("failed to download audio: %v", err)
		return handler.ErrFailedToGetMedia
	}

	_, err = io.Copy(audiofile, stream)
	if err != nil {
		log.Printf("failed to copy audio file: %v", err)
		return handler.ErrFailedToGetMedia
	}

	_, _ = audiofile.Seek(0, 0) // ebal w ryt as well

	fmt.Println(file.Name())
	fmt.Println(audiofile.Name())

	mergedfilepath := os.TempDir() + "/" + randstr.String(32) + ".mp4"
	err = ffmpeg.Output([]*ffmpeg.Stream{ffmpeg.Input(file.Name()), ffmpeg.Input(audiofile.Name())}, mergedfilepath).OverWriteOutput().Run()
	if err != nil {
		log.Printf("failed to merge files: %s", err)
		return handler.ErrFailedToGetMedia
	}

	merged, err := os.Open(mergedfilepath)
	if err != nil {
		log.Printf("failed to open merged file: %s", err)
		return handler.ErrFailedToGetMedia
	}

	inputfile := telego.InputFile{File: merged}

	params := telego.SendVideoParams{
		ChatID: telegoutil.ID(userID),
		Video:  inputfile,
		Width:  videoFormat.Width,
		Height: videoFormat.Height,
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
			Width:  videoFormat.Width,
			Height: videoFormat.Height,
		}); err != nil {
			log.Printf("failed to send tg video to channel %d: %s", c, err)
		}
	}

	mediaFileData := queue.MediaFile{
		SocialNetworkID:   video.ID,
		SocialNetworkName: Code,
		FileID:            tgMsg.Video.FileID,
		FileType:          "video",
		Bot:               h.botname.Name,
	}

	if err := h.db.InsertMediaFile(mediaFileData); err != nil {
		log.Printf("failed to save yt media post download: %s", err)
	}

	return nil
}

func (h Handler) replySingle(video *youtube.Video, videoFormat *youtube.Format, userID int64, client *youtube.Client) error {
	file, err := os.CreateTemp("", randstr.String(32)+".mp4")
	if err != nil {
		log.Printf("failed to create a tmp video file: %v", err)
		return handler.ErrFailedToGetMedia
	}
	defer file.Close()

	// Download the video content
	stream, _, err := client.GetStream(video, videoFormat)
	if err != nil {
		log.Printf("failed to download video: %v", err)
		return handler.ErrFailedToGetMedia
	}

	_, err = io.Copy(file, stream)
	if err != nil {
		log.Printf("failed to copy video file: %v", err)
		return handler.ErrFailedToGetMedia
	}

	defer file.Close()

	inputfile := telego.InputFile{File: file}

	params := telego.SendVideoParams{
		ChatID: telegoutil.ID(userID),
		Video:  inputfile,
		Width:  videoFormat.Width,
		Height: videoFormat.Height,
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
			Width:  videoFormat.Width,
			Height: videoFormat.Height,
		}); err != nil {
			log.Printf("failed to send tg video to channel %d: %s", c, err)
		}
	}

	mediaFileData := queue.MediaFile{
		SocialNetworkID:   video.ID,
		SocialNetworkName: Code,
		FileID:            tgMsg.Video.FileID,
		FileType:          "video",
		Bot:               h.botname.Name,
	}

	if err := h.db.InsertMediaFile(mediaFileData); err != nil {
		log.Printf("failed to save yt media post download: %s", err)
	}

	return nil
}
