package twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func getGuestToken() (int64, error) {
	dest, _ := url.Parse("https://api.x.com/1.1/guest/activate.json")

	client := http.Client{}

	req, _ := http.NewRequest(http.MethodPost, dest.String(), nil) //nolint:noctx
	req.Header.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs=1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get twitter guest token: %w", err)
	}

	defer resp.Body.Close()

	tokenResponse := &struct {
		GuestToken string `json:"guest_token"`
	}{}

	bodyBytes, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(bodyBytes, tokenResponse)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal twitter guest token request: %w", err)
	}

	intval, _ := strconv.ParseInt(tokenResponse.GuestToken, 10, 64)

	return intval, nil
}
