package home

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/common"
)

// openBrowser opens the given URL in the system's default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clockTickMsg:
		m.currentTime = time.Time(msg)
		return m, m.tickClock()

	case booksLoadedMsg:
		m.loading = false
		m.filterPending = false
		m.initialized = true
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.books = msg.books
		if msg.reading != nil {
			m.reading = msg.reading
		}
		if msg.avatarArt != "" {
			m.avatarArt = msg.avatarArt
		}
		items := make([]list.Item, len(m.books))
		for i, ub := range m.books {
			items[i] = bookItem{userBook: ub, filterID: m.filter}
		}
		m.list.SetItems(items)
		return m, nil

	case booksOnlyLoadedMsg:
		m.booksLoading = false
		m.filterPending = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.books = msg.books
		items := make([]list.Item, len(m.books))
		for i, ub := range m.books {
			items[i] = bookItem{userBook: ub, filterID: m.filter}
		}
		m.list.SetItems(items)
		return m, nil

	case filterSettledMsg:
		if msg.filter == m.filter && m.filterPending {
			m.filterPending = false
			m.booksLoading = true
			return m, tea.Batch(m.spinner.Tick, m.loadBooksOnly())
		}
		return m, nil

	case activitiesLoadedMsg:
		m.activityLoading = false
		if msg.err != nil {
			m.activityErr = msg.err
			return m, nil
		}
		m.activityErr = nil
		m.activities = msg.activities
		m.activityCursor = 0
		m.activityScroll = 0
		return m, nil

	case tea.MouseMsg:
		return m, nil

	case spinner.TickMsg:
		if m.loading || m.booksLoading || m.activityLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		if m.activityFocused {
			if m.confirm.Active {
				confirmed, _ := m.confirm.HandleKey(msg.String())
				if !m.confirm.Active {
					if confirmed && m.confirmURL != "" {
						openBrowser(m.confirmURL)
					}
					m.confirmURL = ""
				}
				return m, nil
			}

			k := strings.ToLower(msg.String())
			switch k {
			case "j", "down":
				if m.activityCursor < len(m.activities)-1 {
					m.activityCursor++
				}
				return m, nil
			case "k", "up":
				if m.activityCursor > 0 {
					m.activityCursor--
				}
				return m, nil
			case "enter":
				if m.activityCursor < len(m.activities) {
					act := m.activities[m.activityCursor]
					username := m.user.Username
					if act.User != nil {
						username = act.User.Username
					}
					url := fmt.Sprintf("https://hardcover.app/@%s/activity/%d", username, act.ID)
					m.confirmURL = url
					m.confirm = common.NewConfirm(
						fmt.Sprintf("Open in browser?\n%s", url),
						"open-activity",
					)
				}
				return m, nil
			case "a":
				if m.activityFilter == activityFilterMe {
					m.activityFilter = activityFilterForYou
				} else {
					m.activityFilter = activityFilterMe
				}
				m.activityLoading = true
				m.activityErr = nil
				return m, tea.Batch(m.spinner.Tick, m.loadActivities())
			case "esc":
				m.activityFocused = false
				return m, nil
			}
			return m, nil
		}

		if m.readingFocused {
			k := strings.ToLower(msg.String())
			switch k {
			case "j", "down":
				if m.readingCursor < len(m.reading)-1 {
					m.readingCursor++
				}
				return m, nil
			case "k", "up":
				if m.readingCursor > 0 {
					m.readingCursor--
				}
				return m, nil
			case "enter":
				if m.readingCursor < len(m.reading) {
					ub := m.reading[m.readingCursor]
					return m, func() tea.Msg {
						return NavigateToBookMsg{UserBook: &ub}
					}
				}
			case "esc", "r":
				m.readingFocused = false
				return m, nil
			}
			return m, nil
		}

		if m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		k := strings.ToLower(msg.String())
		switch k {
		case "r":
			if len(m.reading) > 0 {
				m.readingFocused = true
				return m, nil
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(bookItem); ok {
				ub := item.userBook
				return m, func() tea.Msg {
					return NavigateToBookMsg{UserBook: &ub}
				}
			}
		case "f":
			m.filter = (m.filter + 1) % 7
			m.page = 0
			m.filterPending = true
			return m, m.scheduleFilterLoad()
		case "shift+f":
			m.filter = (m.filter + 6) % 7
			m.page = 0
			m.filterPending = true
			return m, m.scheduleFilterLoad()
		case "]":
			if len(m.books) == m.pageSize {
				m.page++
				m.booksLoading = true
				return m, tea.Batch(m.spinner.Tick, m.loadBooksOnly())
			}
		case "[":
			if m.page > 0 {
				m.page--
				m.booksLoading = true
				return m, tea.Batch(m.spinner.Tick, m.loadBooksOnly())
			}
		case "a":
			if len(m.activities) > 0 {
				m.activityFocused = true
				return m, nil
			}
			if m.activityFilter == activityFilterMe {
				m.activityFilter = activityFilterForYou
			} else {
				m.activityFilter = activityFilterMe
			}
			m.activityLoading = true
			m.activityErr = nil
			return m, tea.Batch(m.spinner.Tick, m.loadActivities())

		}
	}

	if !m.loading {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}
