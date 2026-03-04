package tui

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Aayush9029/hn/internal/models"
)

type storyItem struct {
	item models.Item
	rank int
}

func (s storyItem) Title() string {
	domain := s.item.Domain()
	if domain != "" {
		return fmt.Sprintf("%d. %s (%s)", s.rank, s.item.Title, domain)
	}
	return fmt.Sprintf("%d. %s", s.rank, s.item.Title)
}

func (s storyItem) Description() string {
	return fmt.Sprintf("▲ %d  %s  %s  %d comments",
		s.item.Score, s.item.By, s.item.TimeAgo(), s.item.Descendants)
}

func (s storyItem) FilterValue() string { return s.item.Title }

type storiesModel struct {
	list           list.Model
	feed           models.FeedType
	items          []models.Item
	ready          bool
	selectedThread int
	wantRefresh    bool
}

func newStoriesModel(feed models.FeedType, items []models.Item) storiesModel {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = storyItem{item: item, rank: i + 1}
	}

	l := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("🔶 %s", feed.Label())
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Bold(true).
		MarginLeft(2)
	l.SetShowStatusBar(true)
	l.DisableQuitKeybindings()

	return storiesModel{
		list:  l,
		feed:  feed,
		items: items,
		ready: true,
	}
}

func (m storiesModel) update(msg tea.Msg) (storiesModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(storyItem); ok {
				m.selectedThread = item.item.ID
				return m, nil
			}
		case "o":
			if item, ok := m.list.SelectedItem().(storyItem); ok {
				if item.item.URL != "" {
					openBrowser(item.item.URL)
				}
			}
			return m, nil
		case "r":
			m.wantRefresh = true
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m storiesModel) view() string {
	return m.list.View()
}

func (m *storiesModel) setSize(w, h int) {
	if !m.ready {
		return
	}
	m.list.SetSize(w, h)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	_ = cmd.Start()
}
