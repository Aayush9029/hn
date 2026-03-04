package ui

import (
	"html"
	"regexp"
	"strings"
)

var (
	reP       = regexp.MustCompile(`(?i)<p>`)
	reA       = regexp.MustCompile(`(?i)<a\s[^>]*href="([^"]*)"[^>]*>([^<]*)</a>`)
	rePreCode = regexp.MustCompile(`(?is)<pre><code>(.*?)</code></pre>`)
	reTag     = regexp.MustCompile(`<[^>]+>`)
)

// StripHTML converts HN HTML to plain text.
func StripHTML(s string) string {
	if s == "" {
		return ""
	}

	// <pre><code>...</code></pre> → indented code block
	s = rePreCode.ReplaceAllStringFunc(s, func(match string) string {
		inner := rePreCode.FindStringSubmatch(match)
		if len(inner) < 2 {
			return match
		}
		code := html.UnescapeString(inner[1])
		lines := strings.Split(code, "\n")
		var b strings.Builder
		b.WriteString("\n")
		for _, line := range lines {
			b.WriteString("    ")
			b.WriteString(line)
			b.WriteString("\n")
		}
		return b.String()
	})

	// <a href="url">text</a> → text (url)
	s = reA.ReplaceAllString(s, "$2 ($1)")

	// <p> → double newline
	s = reP.ReplaceAllString(s, "\n\n")

	// Strip remaining tags
	s = reTag.ReplaceAllString(s, "")

	// Unescape HTML entities
	s = html.UnescapeString(s)

	return strings.TrimSpace(s)
}

// WordWrap wraps text to the given width.
func WordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var b strings.Builder
	for _, paragraph := range strings.Split(s, "\n") {
		if strings.HasPrefix(paragraph, "    ") {
			// Preserve code block lines
			b.WriteString(paragraph)
			b.WriteString("\n")
			continue
		}
		col := 0
		words := strings.Fields(paragraph)
		for i, w := range words {
			wLen := len(w)
			if col > 0 && col+1+wLen > width {
				b.WriteString("\n")
				col = 0
			}
			if col > 0 {
				b.WriteString(" ")
				col++
			}
			b.WriteString(w)
			col += wLen
			_ = i
		}
		b.WriteString("\n")
	}
	return b.String()
}
