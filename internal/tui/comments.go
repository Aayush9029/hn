package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Aayush9029/hn/internal/models"
	"github.com/Aayush9029/hn/internal/ui"
)

var (
	commentAuthorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)
	commentScoreStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	commentDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	commentHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("208")).
				Bold(true).
				MarginLeft(2)
)

type commentsModel struct {
	viewport    viewport.Model
	story       *models.Item
	comments    []*models.Comment
	ready       bool
	wantRefresh bool
	width       int
	height      int
}

func newCommentsModel(story *models.Item, comments []*models.Comment) commentsModel {
	return commentsModel{
		story:    story,
		comments: comments,
	}
}

func (m *commentsModel) renderContent() string {
	var b strings.Builder
	w := m.width - 4
	if w < 40 {
		w = 40
	}

	// Story header
	domain := m.story.Domain()
	title := m.story.Title
	if domain != "" {
		title = fmt.Sprintf("%s (%s)", title, domain)
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).MarginLeft(2).Render(title))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().MarginLeft(2).Render(
		fmt.Sprintf("%s  %s  %s  %d comments",
			commentScoreStyle.Render(fmt.Sprintf("▲ %d", m.story.Score)),
			commentAuthorStyle.Render(m.story.By),
			commentDimStyle.Render(m.story.TimeAgo()),
			m.story.Descendants,
		),
	))
	b.WriteString("\n")

	// Story self-text
	if m.story.Text != "" {
		text := ui.StripHTML(m.story.Text)
		text = ui.WordWrap(text, w)
		b.WriteString("\n")
		for _, line := range strings.Split(text, "\n") {
			if line != "" {
				b.WriteString("  ")
				b.WriteString(line)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("242")).Render(
		fmt.Sprintf("─── %d comments ───", m.story.Descendants),
	))
	b.WriteString("\n\n")

	m.renderComments(&b, m.comments, w)
	return b.String()
}

func (m *commentsModel) renderComments(b *strings.Builder, comments []*models.Comment, width int) {
	for _, c := range comments {
		m.renderComment(b, c, width)
	}
}

func (m *commentsModel) renderComment(b *strings.Builder, c *models.Comment, width int) {
	indent := strings.Repeat("  ", c.Depth+1)

	// Author + time header
	b.WriteString(indent)
	b.WriteString(commentAuthorStyle.Render(c.Item.By))
	b.WriteString("  ")
	b.WriteString(commentDimStyle.Render(c.Item.TimeAgo()))
	b.WriteString("\n")

	// Comment body
	text := ui.StripHTML(c.Item.Text)
	bodyWidth := width - (c.Depth+1)*2
	if bodyWidth < 20 {
		bodyWidth = 20
	}
	text = ui.WordWrap(text, bodyWidth)
	for _, line := range strings.Split(text, "\n") {
		if line != "" {
			b.WriteString(indent)
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	for _, child := range c.Children {
		m.renderComment(b, child, width)
	}
}

func (m commentsModel) update(msg tea.Msg) (commentsModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "r":
			m.wantRefresh = true
			return m, nil
		case "o":
			if m.story.URL != "" {
				openBrowser(m.story.URL)
			}
			return m, nil
		case "g":
			m.viewport.GotoTop()
			return m, nil
		case "G":
			m.viewport.GotoBottom()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m commentsModel) view() string {
	if !m.ready {
		return ""
	}
	header := commentHeaderStyle.Render("🔶 Thread")
	divider := lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("242")).Render(
		strings.Repeat("─", m.width-4),
	)
	help := lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("242")).Render(
		"j/k scroll • o open • r refresh • esc back • q quit",
	)
	return fmt.Sprintf("%s\n%s\n%s\n%s", header, divider, m.viewport.View(), help)
}

func (m *commentsModel) setSize(w, h int) {
	m.width = w
	m.height = h
	headerHeight := 3 // header + divider + help
	vpHeight := h - headerHeight
	if vpHeight < 1 {
		vpHeight = 1
	}

	if !m.ready && m.story != nil {
		m.viewport = viewport.New(w, vpHeight)
		m.viewport.SetContent(m.renderContent())
		m.ready = true
	} else if m.ready {
		m.viewport.Width = w
		m.viewport.Height = vpHeight
		if m.story != nil {
			m.viewport.SetContent(m.renderContent())
		}
	}
}
