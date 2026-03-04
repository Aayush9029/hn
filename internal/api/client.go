package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/Aayush9029/hn/internal/models"
)

const (
	firebaseURL = "https://hacker-news.firebaseio.com/v0"
	algoliaURL  = "https://hn.algolia.com/api/v1"
)

type Client struct {
	http *http.Client
}

func New() *Client {
	return &Client{
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) get(url string, dest interface{}) error {
	resp, err := c.http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("parsing json: %w", err)
	}
	return nil
}

// FeedIDs fetches the list of story IDs for a given feed type.
func (c *Client) FeedIDs(feed models.FeedType) ([]int, error) {
	var ids []int
	endpoint := fmt.Sprintf("%s/%s.json", firebaseURL, feed.Endpoint())
	if err := c.get(endpoint, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// GetItem fetches a single item by ID.
func (c *Client) GetItem(id int) (*models.Item, error) {
	var item models.Item
	if err := c.get(fmt.Sprintf("%s/item/%d.json", firebaseURL, id), &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// FetchItems fetches multiple items concurrently, preserving order. Skips dead/deleted.
func (c *Client) FetchItems(ids []int, limit int) []models.Item {
	if limit > len(ids) {
		limit = len(ids)
	}
	ids = ids[:limit]

	type result struct {
		idx  int
		item *models.Item
	}

	results := make(chan result, len(ids))
	sem := make(chan struct{}, 20)
	var wg sync.WaitGroup

	for i, id := range ids {
		wg.Add(1)
		go func(idx, id int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			item, err := c.GetItem(id)
			if err != nil || item.Dead || item.Deleted {
				return
			}
			results <- result{idx: idx, item: item}
		}(i, id)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	collected := make([]result, 0, len(ids))
	for r := range results {
		collected = append(collected, r)
	}
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].idx < collected[j].idx
	})

	items := make([]models.Item, 0, len(collected))
	for _, r := range collected {
		items = append(items, *r.item)
	}
	return items
}

// FetchCommentTree recursively fetches and resolves a comment tree.
func (c *Client) FetchCommentTree(kids []int, depth, maxDepth int) []*models.Comment {
	if len(kids) == 0 || depth > maxDepth {
		return nil
	}

	type result struct {
		idx     int
		comment *models.Comment
	}

	results := make(chan result, len(kids))
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for i, id := range kids {
		wg.Add(1)
		go func(idx, id int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			item, err := c.GetItem(id)
			if err != nil || item.Dead || item.Deleted || item.By == "" {
				return
			}
			comment := &models.Comment{
				Item:     *item,
				Depth:    depth,
				Children: c.FetchCommentTree(item.Kids, depth+1, maxDepth),
			}
			results <- result{idx: idx, comment: comment}
		}(i, id)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	collected := make([]result, 0, len(kids))
	for r := range results {
		collected = append(collected, r)
	}
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].idx < collected[j].idx
	})

	comments := make([]*models.Comment, 0, len(collected))
	for _, r := range collected {
		comments = append(comments, r.comment)
	}
	return comments
}

// Search performs an Algolia search for stories.
func (c *Client) Search(query string, limit int) (*models.SearchResponse, error) {
	u := fmt.Sprintf("%s/search?query=%s&tags=story&hitsPerPage=%d",
		algoliaURL, url.QueryEscape(query), limit)
	var resp models.SearchResponse
	if err := c.get(u, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
