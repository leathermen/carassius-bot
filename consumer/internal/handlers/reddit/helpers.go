package reddit

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

var errNoRedditID = errors.New("malformed Reddit link")

var controlSegments = []string{
	"comments",
	"s",
}

func extractRedditID(redditURL string) (string, error) {
	parsedURL, err := url.Parse(redditURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	pathSegments := strings.Split(parsedURL.Path, "/")

	for i, segment := range pathSegments {
		if slices.Contains(controlSegments, segment) && i+1 < len(pathSegments) {
			return pathSegments[i+1], nil
		}
	}

	return "", errNoRedditID
}
