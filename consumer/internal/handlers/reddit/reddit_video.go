package reddit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/gocolly/colly"
	"github.com/grafov/m3u8"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
	"github.com/thanhpk/randstr"
)

var errFailedToGetVideo = errors.New("failed to get Reddit video")

const maxTGFileSize = 1024 * 1024 * 50

func (rh *Handler) video(userID int64, url string) error {
	redditID, err := extractRedditID(url)
	if err != nil {
		return err
	}

	c := colly.NewCollector()

	var (
		success     bool
		errmsg, src string
	)

	c.OnHTML("body", func(e *colly.HTMLElement) {
		shredditScreenview := e.DOM.Find("shreddit-player-2")
		var hasSrc bool
		src, hasSrc = shredditScreenview.Attr("src")
		if !hasSrc {
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

	resp, err := http.Get(src) //nolint:noctx

	if err != nil {
		log.Printf("failed to fetch playlist: %v", err)
		return errFailedToGetVideo
	}

	defer resp.Body.Close()

	playlist, listType, err := m3u8.DecodeFrom(resp.Body, true)
	if err != nil {
		log.Printf("failed to parse playlist: %v", err)
		return errFailedToGetVideo
	}

	if listType != m3u8.MASTER {
		log.Printf("unsupported playlist type: %v (expected MASTER)", listType)
		return errFailedToGetVideo
	}

	masterPL, _ := playlist.(*m3u8.MasterPlaylist)

	var (
		currentMediaPLVar *m3u8.Variant
		// currentAudioPLVar *m3u8.Variant
		currentMPLSize    int64
	)

	for _, mediaPLVar := range masterPL.Variants {
		if mediaPLVar.Audio == "" {
			continue
		}
		fullVariantURL := getFullVariantURL(src, mediaPLVar.URI)
		resp, err := http.Get(fullVariantURL) //nolint:noctx
		if err != nil {
			log.Printf("failed to download reddit video sub playlist: %s", err)
			continue
		}

		defer resp.Body.Close()

		playlist, listType, err := m3u8.DecodeFrom(resp.Body, true)
		if err != nil {
			log.Printf("failed to parse playlist: %v", err)
			continue
		}

		if listType != m3u8.MEDIA {
			continue
		}

		media := playlist.(*m3u8.MediaPlaylist)
		var totalSize int64

		for _, segment := range media.Segments {
			if segment != nil {
				segmentSize, err := getSegmentSize(src, segment.URI)
				if err != nil {
					log.Printf("failed to get segment size: %v", err)
					continue
				}
				totalSize += segmentSize
			}
		}

		if totalSize > currentMPLSize {
			currentMediaPLVar = mediaPLVar
			currentMPLSize = totalSize
		}

	}

	if currentMediaPLVar == nil {
		log.Printf("acceptable reddit video variant is not found")
		return errFailedToGetVideo
	}

	fullVariantURL := getFullVariantURL(src, currentMediaPLVar.URI)
	resp, err = http.Get(fullVariantURL) //nolint:noctx
	if err != nil {
		log.Printf("failed to download reddit video sub playlist: %s", err)
		return errFailedToGetVideo
	}

	defer resp.Body.Close()

	playlist, listType, err = m3u8.DecodeFrom(resp.Body, true)
	if listType != m3u8.MEDIA {
		log.Printf("expected media playlist but got master playlist")
		return errFailedToGetVideo
	}

	mediaPlaylist, _ := playlist.(*m3u8.MediaPlaylist)

	tsfile, err := os.CreateTemp("", "*.ts")
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer tsfile.Close()

	for _, segment := range mediaPlaylist.Segments {
		if segment == nil {
			continue
		}

		segmentURL := resolveURL(src, segment.URI)
		err := appendSegment(segmentURL, tsfile)
		if err != nil {
			return fmt.Errorf("failed to download segment %s: %v", segment.URI, err)
		}

		log.Printf("Appended segment: %s\n", segment.URI) //TODO remove
	}

	mp4filePath := os.TempDir() + randstr.String(24) + ".mp4"
	if err := convertTS2MP4(tsfile.Name(), mp4filePath); err != nil {
		log.Printf("failed to convert ts 2 mp4: %s", err)
		return errFailedToGetVideo
	}

	mp4file, err := os.Open(mp4filePath)

	if err != nil {
		log.Printf("failed to open encoded mp4 file: %s", err)
		return errFailedToGetVideo
	}

	width, height, err := getVideoDimensions(mp4filePath)
	if err != nil {
		log.Printf("failed to get video dimensions: %s", err)
		return errFailedToGetVideo
	}

	tgMsg, err := rh.bot.SendVideo(&telego.SendVideoParams{
		ChatID: telegoutil.ID(userID),
		Video:  telegoutil.File(mp4file),
		Width:  width,
		Height: height,
	})
	if err != nil {
		log.Printf("failed to send user reddit video, user: %d", userID)

		return errFailedToGetVideo
	}

	for _, c := range rh.channels {
		if _, err = rh.bot.SendVideo(&telego.SendVideoParams{
			ChatID: telegoutil.ID(c),
			Video:  telegoutil.FileFromID(tgMsg.Video.FileID),
			Width:  width,
			Height: height,
		}); err != nil {
			log.Printf("failed to send tg video to channel %d: %s", c, err)
		}
	}

	botname, _ := rh.bot.GetMyName(nil)
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

func resolveURL(playlistURL, segmentURI string) string {
	if strings.HasPrefix(segmentURI, "http") {
		return segmentURI
	}
	parsedURL, _ := url.Parse(playlistURL)
	basePath := path.Dir(parsedURL.Path)
	output := fmt.Sprintf("%s://%s%s/%s", parsedURL.Scheme, parsedURL.Host, basePath, segmentURI)
	return output
}

func appendSegment(url string, w io.Writer) error {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	return err
}

func convertTS2MP4(inputTS, outputMP4 string) error {
	cmd := exec.Command("ffmpeg", "-i", inputTS, "-c", "copy", outputMP4)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running ffmpeg: %v", err)
	}

	return nil
}

func getVideoDimensions(inputFile string) (int, int, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=p=0", inputFile)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe error: %v", err)
	}

	dimensions := strings.TrimSpace(out.String())
	parts := strings.Split(dimensions, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected ffprobe output: %s", dimensions)
	}

	var width, height int
	_, _ = fmt.Sscanf(parts[0], "%d", &width)
	_, _ = fmt.Sscanf(parts[1], "%d", &height)

	return width, height, nil
}

func getSegmentSize(baseURL, segmentURI string) (int64, error) {
	segmentURL := baseURL[:strings.LastIndex(baseURL, "/")+1] + segmentURI

	resp, err := http.Head(segmentURL)
	if err != nil {
		return 0, fmt.Errorf("failed to send HEAD request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to retrieve segment size: HTTP status %d", resp.StatusCode)
	}

	return resp.ContentLength, nil
}

func getFullVariantURL(masterURL string, mediaURI string) string {
	// Extract the base URL from the master URL
	baseURL := masterURL[:strings.LastIndex(masterURL, "/")+1]
	return baseURL + mediaURI
}
