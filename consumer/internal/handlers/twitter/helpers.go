package twitter

import "regexp"

func extractTweetID(link string) (string, bool) {
	regex := regexp.MustCompile(`https://.+/[A-Za-z0-9_]+/status/(\d+)`)

	matches := regex.FindStringSubmatch(link)
	if len(matches) != 2 {
		return "", false
	}

	return matches[1], true
}
