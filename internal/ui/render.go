package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/Aayush9029/hn/internal/models"
	"golang.org/x/term"
)

func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// PrintHeader prints the HN branded header.
func PrintHeader() {
	fmt.Printf("\n  %s⚡ hn%s\n", Orange, Reset)
}

// PrintStories prints a list of stories in CLI mode.
func PrintStories(feed models.FeedType, items []models.Item) {
	w := termWidth()
	fmt.Printf("\n  %s%s%s %s(%d stories)%s\n\n", Bold, feed.Label(), Reset, Dim, len(items), Reset)
	for i, item := range items {
		printStory(i+1, item, w)
	}
	fmt.Println()
}

func printStory(n int, item models.Item, width int) {
	domain := item.Domain()
	domainStr := ""
	if domain != "" {
		domainStr = fmt.Sprintf(" %s(%s)%s", Dim, domain, Reset)
	}
	fmt.Printf("  %s%3d.%s %s%s", Orange, n, Reset, item.Title, domainStr)
	fmt.Println()
	fmt.Printf("       %s▲ %d%s  %s%s%s  %s%s  %s%d comments%s\n",
		Orange, item.Score, Reset,
		Author, item.By, Reset,
		Dim, item.TimeAgo(), Reset,
		item.Descendants, Reset,
	)
}

// PrintThread prints a story and its comments in CLI mode.
func PrintThread(story *models.Item, comments []*models.Comment) {
	w := termWidth()

	// Story header
	domain := story.Domain()
	domainStr := ""
	if domain != "" {
		domainStr = fmt.Sprintf(" %s(%s)%s", Dim, domain, Reset)
	}
	fmt.Printf("\n  %s%s%s%s\n", Bold, story.Title, Reset, domainStr)
	fmt.Printf("  %s▲ %d%s  %s%s%s  %s%s  %s%d comments%s\n",
		Orange, story.Score, Reset,
		Author, story.By, Reset,
		Dim, story.TimeAgo(), Reset,
		story.Descendants, Reset,
	)

	if story.Text != "" {
		text := StripHTML(story.Text)
		text = WordWrap(text, w-4)
		for _, line := range strings.Split(text, "\n") {
			if line != "" {
				fmt.Printf("  %s\n", line)
			}
		}
	}
	fmt.Println()

	printComments(comments, w)
	fmt.Println()
}

func printComments(comments []*models.Comment, width int) {
	for _, c := range comments {
		printComment(c, width)
	}
}

func printComment(c *models.Comment, width int) {
	indent := strings.Repeat("  ", c.Depth)
	fmt.Printf("%s  %s%s%s  %s%s%s\n", indent, Author, c.Item.By, Reset, Dim, c.Item.TimeAgo(), Reset)

	text := StripHTML(c.Item.Text)
	bodyWidth := width - (c.Depth*2 + 2)
	if bodyWidth < 20 {
		bodyWidth = 20
	}
	text = WordWrap(text, bodyWidth)
	for _, line := range strings.Split(text, "\n") {
		if line != "" {
			fmt.Printf("%s  %s\n", indent, line)
		}
	}
	fmt.Println()

	for _, child := range c.Children {
		printComment(child, width)
	}
}

// PrintSearch prints Algolia search results in CLI mode.
func PrintSearch(query string, resp *models.SearchResponse) {
	fmt.Printf("\n  %sSearch: %s%s %s(%d results)%s\n\n", Bold, query, Reset, Dim, resp.NbHits, Reset)
	for i, hit := range resp.Hits {
		domain := hit.Domain()
		domainStr := ""
		if domain != "" {
			domainStr = fmt.Sprintf(" %s(%s)%s", Dim, domain, Reset)
		}
		fmt.Printf("  %s%3d.%s %s%s\n", Orange, i+1, Reset, hit.Title, domainStr)
		fmt.Printf("       %s▲ %d%s  %s%s%s  %s%d comments%s\n",
			Orange, hit.Points, Reset,
			Author, hit.Author, Reset,
			Dim, hit.NumComments, Reset,
		)
	}
	fmt.Println()
}
