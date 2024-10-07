package insta

import "regexp"

func getStoryID(url string) (string, bool) {
	pattern := regexp.MustCompile(`(?i)/stories/([A-Za-z0-9._]+)/(\d+)`)
	if match := pattern.FindStringSubmatch(url); len(match) > 2 {
		return match[2], true
	}

	return "", false
}
