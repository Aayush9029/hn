package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Aayush9029/hn/internal/api"
	"github.com/Aayush9029/hn/internal/models"
)

type viewState int

const (
	viewFeeds viewState = iota
	viewStories
	viewComments
)

type model struct {
	state       viewState
	client      *api.Client
	feeds       feedsModel
	stories     storiesModel
	comments    commentsModel
	width       int
	height      int
	loading     bool
	loadingMsg  string
	quitting    bool
	countPrefix int
}

// Async messages
type feedLoadedMsg struct {
	feed  models.FeedType
	items []models.Item
}

type threadLoadedMsg struct {
	story    *models.Item
	comments []*models.Comment
}

type errMsg struct{ err error }

func initialModel() model {
	return model{
		state:  viewFeeds,
		client: api.New(),
		feeds:  newFeedsModel(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) fetchFeed(feed models.FeedType) tea.Cmd {
	return func() tea.Msg {
		ids, err := m.client.FeedIDs(feed)
		if err != nil {
			return errMsg{err}
		}
		items := m.client.FetchItems(ids, 30)
		return feedLoadedMsg{feed: feed, items: items}
	}
}

func (m model) fetchThread(id int) tea.Cmd {
	return func() tea.Msg {
		story, err := m.client.GetItem(id)
		if err != nil {
			return errMsg{err}
		}
		comments := m.client.FetchCommentTree(story.Kids, 0, 5)
		return threadLoadedMsg{story: story, comments: comments}
	}
}

func (m model) isFiltering() bool {
	switch m.state {
	case viewFeeds:
		return m.feeds.list.FilterState() == list.Filtering
	case viewStories:
		return m.stories.list.FilterState() == list.Filtering
	}
	return false
}

func (m model) hasFilterApplied() bool {
	switch m.state {
	case viewFeeds:
		return m.feeds.list.FilterState() == list.FilterApplied
	case viewStories:
		return m.stories.list.FilterState() == list.FilterApplied
	}
	return false
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		if key == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		if !m.isFiltering() {
			// Accumulate vim count prefix
			if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
				digit := int(key[0] - '0')
				if m.countPrefix == 0 && digit == 0 {
					// 0 alone doesn't accumulate
				} else {
					m.countPrefix = m.countPrefix*10 + digit
					return m, nil
				}
			}

			count := m.countPrefix
			if count < 1 {
				count = 1
			}
			isVimMotion := key == "j" || key == "k" || key == "G" || key == "g"

			if isVimMotion && m.countPrefix > 0 {
				m.countPrefix = 0
				var cmds []tea.Cmd
				for i := 0; i < count; i++ {
					var cmd tea.Cmd
					switch m.state {
					case viewFeeds:
						m.feeds, cmd = m.feeds.update(msg)
					case viewStories:
						m.stories, cmd = m.stories.update(msg)
					case viewComments:
						m.comments, cmd = m.comments.update(msg)
					}
					if cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
				// Check selection after replay
				switch m.state {
				case viewFeeds:
					if m.feeds.chosen {
						feed := m.feeds.selected
						m.feeds.chosen = false
						m.loading = true
						m.loadingMsg = fmt.Sprintf("Loading %s...", feed.Label())
						return m, m.fetchFeed(feed)
					}
				case viewStories:
					if m.stories.selectedThread != 0 {
						id := m.stories.selectedThread
						m.stories.selectedThread = 0
						m.loading = true
						m.loadingMsg = "Loading thread..."
						return m, m.fetchThread(id)
					}
				}
				if len(cmds) > 0 {
					return m, tea.Batch(cmds...)
				}
				return m, nil
			}

			m.countPrefix = 0

			switch key {
			case "q":
				m.quitting = true
				return m, tea.Quit
			case "esc":
				if !m.hasFilterApplied() {
					switch m.state {
					case viewComments:
						m.state = viewStories
						return m, nil
					case viewStories:
						m.state = viewFeeds
						return m, nil
					case viewFeeds:
						m.quitting = true
						return m, tea.Quit
					}
				}
			case "backspace":
				switch m.state {
				case viewComments:
					m.state = viewStories
					return m, nil
				case viewStories:
					m.state = viewFeeds
					return m, nil
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.feeds.setSize(msg.Width, msg.Height)
		m.stories.setSize(msg.Width, msg.Height)
		m.comments.setSize(msg.Width, msg.Height)

	case feedLoadedMsg:
		m.loading = false
		m.stories = newStoriesModel(msg.feed, msg.items)
		m.stories.setSize(m.width, m.height)
		m.state = viewStories
		return m, nil

	case threadLoadedMsg:
		m.loading = false
		m.comments = newCommentsModel(msg.story, msg.comments)
		m.comments.setSize(m.width, m.height)
		m.state = viewComments
		return m, nil

	case errMsg:
		m.loading = false
		m.loadingMsg = fmt.Sprintf("Error: %s — press q to quit", msg.err)
		return m, nil
	}

	// Delegate to current view
	var cmd tea.Cmd
	switch m.state {
	case viewFeeds:
		m.feeds, cmd = m.feeds.update(msg)
		if m.feeds.chosen {
			feed := m.feeds.selected
			m.feeds.chosen = false
			m.loading = true
			m.loadingMsg = fmt.Sprintf("Loading %s...", feed.Label())
			return m, m.fetchFeed(feed)
		}
	case viewStories:
		m.stories, cmd = m.stories.update(msg)
		if m.stories.selectedThread != 0 {
			id := m.stories.selectedThread
			m.stories.selectedThread = 0
			m.loading = true
			m.loadingMsg = "Loading thread..."
			return m, m.fetchThread(id)
		}
		if m.stories.wantRefresh {
			m.stories.wantRefresh = false
			m.loading = true
			m.loadingMsg = fmt.Sprintf("Refreshing %s...", m.stories.feed.Label())
			return m, m.fetchFeed(m.stories.feed)
		}
	case viewComments:
		m.comments, cmd = m.comments.update(msg)
		if m.comments.wantRefresh {
			m.comments.wantRefresh = false
			m.loading = true
			m.loadingMsg = "Refreshing thread..."
			return m, m.fetchThread(m.comments.story.ID)
		}
	}

	return m, cmd
}

var loadingStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("208")).
	Bold(true).
	MarginLeft(2).
	MarginTop(1)

func (m model) View() string {
	if m.quitting {
		return ""
	}
	if m.loading {
		return loadingStyle.Render("⏳ " + m.loadingMsg)
	}

	switch m.state {
	case viewFeeds:
		return m.feeds.view()
	case viewStories:
		return m.stories.view()
	case viewComments:
		return m.comments.view()
	}
	return ""
}

func Run() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
