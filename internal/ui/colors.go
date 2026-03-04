package ui

import "os"

var (
	Green  = "\033[1;32m"
	Red    = "\033[1;31m"
	Yellow = "\033[1;33m"
	Cyan   = "\033[1;36m"
	Blue   = "\033[1;34m"
	Dim    = "\033[2m"
	Bold   = "\033[1m"
	Reset  = "\033[0m"

	// HN-flavored
	Orange = "\033[38;5;208m" // HN orange for scores/headers
	Author = "\033[1;32m"     // green for usernames
	Domain = "\033[2m"        // dim for domains
)

func init() {
	if !isTerminal() {
		Green = ""
		Red = ""
		Yellow = ""
		Cyan = ""
		Blue = ""
		Dim = ""
		Bold = ""
		Reset = ""
		Orange = ""
		Author = ""
		Domain = ""
	}
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
