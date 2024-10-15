package youtube

import (
	"regexp"
)

func extractShortVideoID(link string) (string, bool) {
	regex := regexp.MustCompile(`(?:youtu\.be/|/v/|/u/\w/|/embed/|/shorts/|/e/|/at/|/vi/|watch\?v=|/m/)([^/?&]+)`)

	matches := regex.FindStringSubmatch(link)
	if len(matches) != 2 {
		return "", false
	}

	return matches[1], true
}
