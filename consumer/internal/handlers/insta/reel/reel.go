package reel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

var dest, _ = url.Parse("https://www.instagram.com/graphql/query")

type Query struct {
	A                    string `json:"__a"`
	Ccg                  string `json:"__ccg"`
	CometReq             string `json:"__comet_req"`
	Csr                  string `json:"__csr"`
	D                    string `json:"__d"`
	Dyn                  string `json:"__dyn"`
	Hs                   string `json:"__hs"`
	Hsi                  string `json:"__hsi"`
	Req                  string `json:"__req"`
	Rev                  string `json:"__rev"`
	S                    string `json:"__s"`
	SpinB                string `json:"__spin_b"`
	SpinR                string `json:"__spin_r"`
	SpinT                string `json:"__spin_t"`
	User                 string `json:"__user"`
	Av                   string `json:"av"`
	DocID                string `json:"doc_id"`
	Dpr                  string `json:"dpr"`
	FBApiCallerClass     string `json:"fb_api_caller_class"`
	FBApiReqFriendlyName string `json:"fb_api_req_friendly_name"`
	Jazoest              string `json:"jazoest"`
	Lsd                  string `json:"lsd"`
	ServerTimestamps     string `json:"server_timestamps"`
	Variables            string `json:"variables"`
}

type Details struct {
	Data struct {
		Media struct {
			Dimensions struct {
				Height int `json:"height"`
				Width  int `json:"width"`
			} `json:"dimensions"`
			VideoURL string `json:"video_url"`
		} `json:"xdt_shortcode_media"`
	} `json:"data"`
}

func newQuery(reelID string) *Query {
	return &Query{
		A:                    "1",
		Ccg:                  "UNKNOWN",
		CometReq:             "7",
		Csr:                  "iMoMDbsn9_KGAExqGQXllhFaVaHZdbJ5KnGiEF5BOk8SeKSKt9oybCBmQiFvGEKmAaGmibiFGtBKmfzQGB-EGUSrGt7y4uVqK8z8HhU8USKSq6-q8gsWx6vGiqmaWUmgzAh4m7UgxmbCGjw04V7Bwoqx56iGGw2hEiw7xwcuOjwlo2cxeu0IoaU0nBw7Uxgw0XS0MVNcE2xa8wayqNhA94wb21Lgnwj99z0Zz4oEbS6E2Vwp98qG1og9Sq0xAew4kxe17L6JgyE0bMo04X205PE",
		D:                    "www",
		Dyn:                  "7xeUjG1mxu1syUbFp41twpUnwgU7SbzEdF8aUco2qwJw5ux609vCwjE1xoswaq0yE462mcw5Mx62G5UswoEcE7O2l0Fwqo31w9O1TwQzXwae4UaEW2G0AEco5G0zK5o4q0HUvw5rwSyES1TwVwDwHg2ZwrUdUbGwmk0zU8oC1Iwqo5q3e3zhA6bwIxe6V89F8uwm9EO2e2e0N9Wy8Cu",
		Hs:                   "20008.HYP:instagram_web_pkg.2.1..0.0",
		Hsi:                  "7425025843897442975",
		Req:                  "7",
		Rev:                  "1017300160",
		S:                    "uy3c1d:kci9oe:obyzwb",
		SpinB:                "trunk",
		SpinR:                "1017300160",
		SpinT:                "1728773546",
		User:                 "0",
		Av:                   "0",
		DocID:                "8845758582119845",
		Dpr:                  "2",
		FBApiCallerClass:     "RelayModern",
		FBApiReqFriendlyName: "PolarisPostActionLoadPostQueryQuery",
		Jazoest:              "21023",
		Lsd:                  "AVqxQmvKGFs",
		ServerTimestamps:     "true",
		Variables:            "{\"shortcode\":\"" + reelID + "\",\"fetch_tagged_user_count\":null,\"hoisted_comment_id\":null,\"hoisted_reply_id\":null}",
	}
}

func GetURL(reelID, csrf string) (*Details, error) {
	query := newQuery(reelID)

	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal insta query: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, dest.String(), bytes.NewBuffer(jsonData)) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("failed to create insta request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create insta cookie jar: %w", err)
	}

	jar.SetCookies(dest, []*http.Cookie{
		{
			Name:    "csrftoken",
			Value:   csrf,
			Expires: time.Now().Add(time.Hour * 24 * 31),
		},
	})

	client := &http.Client{
		Jar: jar,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query insta graphql: %w", err)
	}

	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	respData := &Details{}
	err = json.Unmarshal(bodyBytes, respData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall insta graphql response: %w", err)
	}

	return respData, nil
}
