package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Aayush9029/hn/internal/models"
)

type feedItem struct {
	feed models.FeedType
}

func (f feedItem) Title() string       { return f.feed.Label() }
func (f feedItem) Description() string { return f.feed.Description() }
func (f feedItem) FilterValue() string { return f.feed.Label() }

type feedsModel struct {
	list     list.Model
	selected models.FeedType
	chosen   bool
}

func newFeedsModel() feedsModel {
	items := make([]list.Item, 0, 6)
	for _, f := range models.AllFeeds() {
		items = append(items, feedItem{feed: f})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "🔶 Hacker News"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Bold(true).
		MarginLeft(2)
	l.SetShowStatusBar(false)
	l.DisableQuitKeybindings()

	return feedsModel{list: l}
}

func (m feedsModel) update(msg tea.Msg) (feedsModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			if item, ok := m.list.SelectedItem().(feedItem); ok {
				m.selected = item.feed
				m.chosen = true
				return m, nil
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m feedsModel) view() string {
	return m.list.View()
}

func (m *feedsModel) setSize(w, h int) {
	m.list.SetSize(w, h)
}
