package insta

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type InstagramRequest struct {
	QueryHash string
	ContentID string
	IsStory   bool
}

func makeInstagramRequest(req *InstagramRequest, cookie string) (string, error) {
	var url string
	fmt.Println(req.IsStory)
	if req.IsStory {
		url = fmt.Sprintf("https://i.instagram.com/api/v1/media/%s/info/", req.ContentID)
	} else {
		url = fmt.Sprintf("https://www.instagram.com/graphql/query/?query_hash=%s&shortcode=%s", req.QueryHash, req.ContentID)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Host", "www.instagram.com")
	httpReq.Header.Set("User-Agent", "Instagram 315.0.0.29.109 Android (33/13; 1980dpi; TCL; T610K; Model_3; mt6765; es_ES; 558601268)")
	//httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:126.0) Gecko/20100101 Firefox/126.0)")
	httpReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	httpReq.Header.Set("Accept-Language", "ru,en-US;q=0.7,en;q=0.3")
	httpReq.Header.Set("Accept-Encoding", "deflate")
	httpReq.Header.Set("DNT", "1")
	httpReq.Header.Set("Connection", "keep-alive")
	httpReq.Header.Set("Cookie", cookie)
	httpReq.Header.Set("Upgrade-Insecure-Requests", "1")
	httpReq.Header.Set("Sec-Fetch-Dest", "document")
	httpReq.Header.Set("Sec-Fetch-Mode", "navigate")
	httpReq.Header.Set("Sec-Fetch-Site", "none")
	httpReq.Header.Set("Sec-Fetch-User", "?1")
	httpReq.Header.Set("Sec-GPC", "1")
	httpReq.Header.Set("Pragma", "no-cache")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("TE", "trailers")

	client := http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func downloadInstagramMedia(mediaURL, filename, cookie string) error {
	client := &http.Client{}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mediaURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Host", "www.instagram.com")
	req.Header.Set("User-Agent", "Instagram 315.0.0.29.109 Android (33/13; 1980dpi; TCL; T610K; Model_3; mt6765; es_ES; 558601268)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru,en-US;q=0.7,en;q=0.3")
	req.Header.Set("Accept-Encoding", "deflate")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("TE", "trailers")

	req.Header.Set("Cookie", cookie)

	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("HTTP-запрос завершился неудачно с кодом статуса: %d", response.StatusCode)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}
