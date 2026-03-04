package models

import (
	"fmt"
	"net/url"
	"time"
)

// FeedType represents a Hacker News feed category.
type FeedType int

const (
	FeedTop FeedType = iota
	FeedNew
	FeedBest
	FeedAsk
	FeedShow
	FeedJobs
)

var feedTypes = []FeedType{FeedTop, FeedNew, FeedBest, FeedAsk, FeedShow, FeedJobs}

func AllFeeds() []FeedType { return feedTypes }

func (f FeedType) Endpoint() string {
	switch f {
	case FeedTop:
		return "topstories"
	case FeedNew:
		return "newstories"
	case FeedBest:
		return "beststories"
	case FeedAsk:
		return "askstories"
	case FeedShow:
		return "showstories"
	case FeedJobs:
		return "jobstories"
	}
	return "topstories"
}

func (f FeedType) Label() string {
	switch f {
	case FeedTop:
		return "Top"
	case FeedNew:
		return "New"
	case FeedBest:
		return "Best"
	case FeedAsk:
		return "Ask HN"
	case FeedShow:
		return "Show HN"
	case FeedJobs:
		return "Jobs"
	}
	return "Top"
}

func (f FeedType) Description() string {
	switch f {
	case FeedTop:
		return "Top stories on Hacker News"
	case FeedNew:
		return "Newest stories"
	case FeedBest:
		return "Best stories of all time"
	case FeedAsk:
		return "Ask HN discussions"
	case FeedShow:
		return "Show HN projects"
	case FeedJobs:
		return "Job postings"
	}
	return ""
}

// FeedTypeFromString parses a feed name string.
func FeedTypeFromString(s string) (FeedType, bool) {
	switch s {
	case "top":
		return FeedTop, true
	case "new":
		return FeedNew, true
	case "best":
		return FeedBest, true
	case "ask":
		return FeedAsk, true
	case "show":
		return FeedShow, true
	case "jobs":
		return FeedJobs, true
	}
	return FeedTop, false
}

// Item represents a Hacker News item (story, comment, job, poll, etc).
type Item struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Time        int64  `json:"time"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Text        string `json:"text"`
	Score       int    `json:"score"`
	Kids        []int  `json:"kids"`
	Descendants int    `json:"descendants"`
	Dead        bool   `json:"dead"`
	Deleted     bool   `json:"deleted"`
}

// Domain extracts the hostname from the item URL.
func (it *Item) Domain() string {
	if it.URL == "" {
		return ""
	}
	u, err := url.Parse(it.URL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// TimeAgo returns a human-readable relative time string.
func (it *Item) TimeAgo() string {
	return TimeAgo(it.Time)
}

// Comment is a resolved tree node for nested comment display.
type Comment struct {
	Item     Item
	Children []*Comment
	Depth    int
}

// SearchHit is a single result from the Algolia search API.
type SearchHit struct {
	ObjectID    string `json:"objectID"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Author      string `json:"author"`
	Points      int    `json:"points"`
	NumComments int    `json:"num_comments"`
	CreatedAt   string `json:"created_at"`
	StoryText   string `json:"story_text"`
}

// Domain extracts the hostname from the search hit URL.
func (h *SearchHit) Domain() string {
	if h.URL == "" {
		return ""
	}
	u, err := url.Parse(h.URL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// SearchResponse is the top-level Algolia search response.
type SearchResponse struct {
	Hits    []SearchHit `json:"hits"`
	NbHits  int         `json:"nbHits"`
	NbPages int         `json:"nbPages"`
}

// TimeAgo formats a Unix timestamp as a relative time string.
func TimeAgo(unix int64) string {
	t := time.Unix(unix, 0)
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	default:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}
