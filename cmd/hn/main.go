package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Aayush9029/hn/internal/api"
	"github.com/Aayush9029/hn/internal/models"
	"github.com/Aayush9029/hn/internal/tui"
	"github.com/Aayush9029/hn/internal/ui"
)

const version = "0.1.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		showHelp()
		os.Exit(0)
	}

	cmd := args[0]

	switch cmd {
	case "-h", "--help", "help":
		showHelp()

	case "-v", "--version", "version":
		fmt.Printf("hn %s\n", version)

	case "-i", "--interactive", "interactive":
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s\n", err)
			os.Exit(1)
		}

	case "top", "new", "best", "ask", "show", "jobs":
		feed, _ := models.FeedTypeFromString(cmd)
		limit := 30
		if len(args) >= 2 {
			if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
				limit = n
			}
		}
		cmdFeed(feed, limit)

	case "thread", "t":
		if len(args) < 2 {
			ui.PrintHeader()
			fmt.Printf("\n%s✗ Usage: hn thread <id>%s\n", ui.Red, ui.Reset)
			fmt.Printf("  %sExample: hn thread 12345678%s\n\n", ui.Dim, ui.Reset)
			os.Exit(1)
		}
		id, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("%s✗ Invalid thread ID: %s%s\n", ui.Red, args[1], ui.Reset)
			os.Exit(1)
		}
		cmdThread(id)

	case "search", "s":
		if len(args) < 2 {
			ui.PrintHeader()
			fmt.Printf("\n%s✗ Usage: hn search <query>%s\n", ui.Red, ui.Reset)
			fmt.Printf("  %sExample: hn search \"rust lang\"%s\n\n", ui.Dim, ui.Reset)
			os.Exit(1)
		}
		query := args[1]
		limit := 20
		if len(args) >= 3 {
			if n, err := strconv.Atoi(args[2]); err == nil && n > 0 {
				limit = n
			}
		}
		cmdSearch(query, limit)

	default:
		fmt.Printf("%s✗ Unknown command: %s%s\n", ui.Red, cmd, ui.Reset)
		fmt.Printf("  %sRun 'hn --help' for usage%s\n", ui.Dim, ui.Reset)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println()
	ui.PrintHeader()
	fmt.Printf("  %sbrowse Hacker News from the terminal%s\n", ui.Dim, ui.Reset)
	fmt.Println()
	fmt.Printf("%sUSAGE%s\n", ui.Blue, ui.Reset)
	fmt.Printf("    hn <command> [args]\n")
	fmt.Println()
	fmt.Printf("%sCOMMANDS%s\n", ui.Blue, ui.Reset)
	fmt.Printf("    -i, --interactive              Launch interactive TUI\n")
	fmt.Printf("    top/new/best/ask/show/jobs [n]  List stories (default 30)\n")
	fmt.Printf("    thread, t <id>                 View story + comments\n")
	fmt.Printf("    search, s <query> [n]          Search stories (default 20)\n")
	fmt.Println()
	fmt.Printf("%sEXAMPLES%s\n", ui.Blue, ui.Reset)
	fmt.Printf("    %shn -i%s                  %s# interactive TUI mode%s\n", ui.Dim, ui.Reset, ui.Dim, ui.Reset)
	fmt.Printf("    %shn top 10%s              %s# top 10 stories%s\n", ui.Dim, ui.Reset, ui.Dim, ui.Reset)
	fmt.Printf("    %shn thread 12345678%s     %s# read a thread%s\n", ui.Dim, ui.Reset, ui.Dim, ui.Reset)
	fmt.Printf("    %shn search \"rust lang\"%s  %s# search stories%s\n", ui.Dim, ui.Reset, ui.Dim, ui.Reset)
	fmt.Println()
	fmt.Printf("%sVERSION%s\n", ui.Blue, ui.Reset)
	fmt.Printf("    %s\n", version)
	fmt.Println()
}

func cmdFeed(feed models.FeedType, limit int) {
	client := api.New()
	ids, err := client.FeedIDs(feed)
	if err != nil {
		fmt.Printf("%s✗ Failed to fetch %s stories: %s%s\n", ui.Red, feed.Label(), err, ui.Reset)
		os.Exit(1)
	}
	ui.PrintHeader()
	items := client.FetchItems(ids, limit)
	ui.PrintStories(feed, items)
}

func cmdThread(id int) {
	client := api.New()
	story, err := client.GetItem(id)
	if err != nil {
		fmt.Printf("%s✗ Failed to fetch item %d: %s%s\n", ui.Red, id, err, ui.Reset)
		os.Exit(1)
	}
	ui.PrintHeader()
	comments := client.FetchCommentTree(story.Kids, 0, 5)
	ui.PrintThread(story, comments)
}

func cmdSearch(query string, limit int) {
	client := api.New()
	resp, err := client.Search(query, limit)
	if err != nil {
		fmt.Printf("%s✗ Search failed: %s%s\n", ui.Red, err, ui.Reset)
		os.Exit(1)
	}
	ui.PrintHeader()
	ui.PrintSearch(query, resp)
}
