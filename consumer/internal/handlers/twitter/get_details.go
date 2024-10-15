package twitter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type twitterMediaType int

const (
	twitterMediaTypePhoto twitterMediaType = iota
	twitterMediaTypeVideo twitterMediaType = iota
	twitterMediaTypeGif   twitterMediaType = iota
)

var errNoEntities = errors.New("no twitter entities")

type tweetDetails struct {
	typ    twitterMediaType
	url    string
	height int
	width  int
}

const (
	twurlbase     = `https://api.x.com/graphql/OoJd6A50cv8GsifjoOHGfg/TweetResultByRestId?variables=%s&features=%s`
	twurlvars     = `{"tweetId":"%s","withCommunity":false,"includePromotedContent":false,"withVoice":false}`
	twurlfeatures = `{"creator_subscriptions_tweet_preview_api_enabled":true,"communities_web_enable_tweet_community_results_fetch":true,"c9s_tweet_anatomy_moderator_badge_enabled":true,"articles_preview_enabled":true,"responsive_web_edit_tweet_api_enabled":true,"graphql_is_translatable_rweb_tweet_is_translatable_enabled":true,"view_counts_everywhere_api_enabled":true,"longform_notetweets_consumption_enabled":true,"responsive_web_twitter_article_tweet_consumption_enabled":true,"tweet_awards_web_tipping_enabled":false,"creator_subscriptions_quote_tweet_preview_enabled":false,"freedom_of_speech_not_reach_fetch_enabled":true,"standardized_nudges_misinfo":true,"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled":true,"rweb_video_timestamps_enabled":true,"longform_notetweets_rich_text_read_enabled":true,"longform_notetweets_inline_media_enabled":true,"rweb_tipjar_consumption_enabled":true,"responsive_web_graphql_exclude_directive_enabled":true,"verified_phone_label_enabled":false,"responsive_web_graphql_skip_user_profile_image_extensions_enabled":false,"responsive_web_graphql_timeline_navigation_enabled":true,"responsive_web_enhance_cards_enabled":false}&fieldToggles={"withArticleRichContentState":true,"withArticlePlainText":false,"withGrokAnalyze":false,"withDisallowedReplyControls":false}`
)

type TweetResult struct {
	Data struct {
		TweetResult struct {
			Result struct {
				Legacy struct {
					Entities struct {
						Media []struct {
							Type          string `json:"type"`
							MediaURLHTTPS string `json:"media_url_https"`
							OriginalInfo  struct {
								Height int `json:"height"`
								Width  int `json:"width"`
							} `json:"original_info"`
							VideoInfo struct {
								Variants []struct {
									Bitrate     int    `json:"bitrate"`
									ContentType string `json:"content_type"`
									URL         string `json:"url"`
								} `json:"variants"`
							} `json:"video_info"`
						} `json:"media"`
					} `json:"entities"`
				} `json:"legacy"`
			} `json:"result"`
		} `json:"tweetResult"`
	} `json:"data"`
}

func getTweetDetails(twid string, guestToken int64) (*tweetDetails, error) {
	fullurl := fmt.Sprintf(twurlbase, url.QueryEscape(fmt.Sprintf(twurlvars, twid)), url.QueryEscape(twurlfeatures))
	dest, _ := url.Parse(fullurl)
	client := http.Client{}

	req, _ := http.NewRequest(http.MethodGet, dest.String(), nil) //nolint:noctx
	req.Header.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs=1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Set("X-Guest-Token", strconv.FormatInt(guestToken, 10))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to http get twitter: %w", err)
	}

	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	twres := TweetResult{}

	err = json.Unmarshal(bodyBytes, &twres)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal twitter response: %w", err)
	}

	if len(twres.Data.TweetResult.Result.Legacy.Entities.Media) == 0 {
		return nil, errNoEntities
	}

	entity := twres.Data.TweetResult.Result.Legacy.Entities.Media[0]

	details := &tweetDetails{
		height: entity.OriginalInfo.Height,
		width:  entity.OriginalInfo.Width,
	}

	switch entity.Type {
	case "photo":
		details.typ = twitterMediaTypePhoto
		details.url = entity.MediaURLHTTPS
	case "video":
		details.typ = twitterMediaTypeVideo
		variant := entity.VideoInfo.Variants[0]
		for _, v := range entity.VideoInfo.Variants {
			if v.Bitrate > variant.Bitrate {
				variant = v
			}
		}
		details.url = variant.URL
	case "animated_gif":
		details.typ = twitterMediaTypeGif
		variant := entity.VideoInfo.Variants[0]
		for _, v := range entity.VideoInfo.Variants {
			if v.Bitrate > variant.Bitrate {
				variant = v
			}
		}
		details.url = variant.URL
	}

	return details, nil
}
