package progress

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	datepicker "github.com/ethanefung/bubble-datepicker"

	"hardcover-tui/internal/api"
	"hardcover-tui/internal/api/mutations"
	"hardcover-tui/internal/common"
)

type progressUpdatedMsg struct {
	err error
}

// NavigateBackMsg signals the app to pop back from the progress screen.
type NavigateBackMsg struct{}

type focusField int

const (
	focusPage focusField = iota
	focusStarted
	focusFinished
)

// Model is the progress update screen model.
type Model struct {
	client         *api.Client
	user           *api.User
	userBook       *api.UserBook
	pageInput      textinput.Model
	startedPicker  datepicker.Model
	finishedPicker datepicker.Model
	focus          focusField
	spinner        spinner.Model
	loading        bool
	err            error
	success        bool
	width          int
	height         int
	confirming      bool
	pendingPages    int
	pendingStarted  *string
	pendingFinished *string
}

// New creates a new progress screen.
func New(client *api.Client, user *api.User, ub *api.UserBook) *Model {
	ti := textinput.New()
	ti.Placeholder = "Current page..."
	ti.Width = 20
	ti.Cursor.Style = common.CursorStyle
	ti.Focus()

	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)

	sp := datepicker.New(time.Now())
	sp.Focused = datepicker.FocusNone
	fp := datepicker.New(time.Now())
	fp.Focused = datepicker.FocusNone

	m := &Model{
		client:         client,
		user:           user,
		userBook:       ub,
		pageInput:      ti,
		startedPicker:  sp,
		finishedPicker: fp,
		focus:          focusPage,
		spinner:        s,
	}

	if ub != nil && len(ub.UserBookReads) > 0 {
		read := ub.UserBookReads[0]
		if read.ProgressPages != nil {
			ti.SetValue(fmt.Sprintf("%d", *read.ProgressPages))
		}
		if read.StartedAt != nil {
			if t, err := time.Parse("2006-01-02", *read.StartedAt); err == nil {
				m.startedPicker.SetTime(t)
				m.startedPicker.SelectDate()
			}
		}
		if read.FinishedAt != nil {
			if t, err := time.Parse("2006-01-02", *read.FinishedAt); err == nil {
				m.finishedPicker.SetTime(t)
				m.finishedPicker.SelectDate()
			}
		}
	}
	m.pageInput = ti

	return m
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

// SetSize updates the available terminal dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// InputFocused returns true when an input is active.
func (m *Model) InputFocused() bool {
	return !m.loading && !m.success && !m.confirming
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case progressUpdatedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		m.success = true
		return m, tea.Batch(
			common.NotifyCmd(common.NotifySuccess, "Progress updated"),
			func() tea.Msg { return NavigateBackMsg{} },
		)

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		if m.success {
			if msg.String() == "esc" {
				return m, func() tea.Msg { return NavigateBackMsg{} }
			}
			return m, nil
		}

		if m.confirming {
			switch msg.String() {
			case "y", "Y":
				m.confirming = false
				m.loading = true
				m.err = nil
				return m, tea.Batch(m.spinner.Tick, m.updateProgress(m.pendingPages, m.pendingStarted, m.pendingFinished))
			case "n", "N", "esc":
				m.confirming = false
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return NavigateBackMsg{} }
		case "tab":
			switch m.focus {
			case focusPage:
				m.focus = focusStarted
				m.pageInput.Blur()
				m.startedPicker.SetFocus(datepicker.FocusCalendar)
				m.finishedPicker.Focused = datepicker.FocusNone
			case focusStarted:
				m.focus = focusFinished
				m.startedPicker.Focused = datepicker.FocusNone
				m.finishedPicker.SetFocus(datepicker.FocusCalendar)
			case focusFinished:
				m.focus = focusPage
				m.finishedPicker.Focused = datepicker.FocusNone
				m.pageInput.Focus()
			}
			return m, nil

		case "enter":
			if m.focus == focusPage {
				pagesStr := strings.TrimSpace(m.pageInput.Value())
				pages, err := strconv.Atoi(pagesStr)
				if err != nil || pages < 0 {
					m.err = fmt.Errorf("please enter a valid page number")
					return m, nil
				}
				m.err = nil

				var startedAt, finishedAt *string
				if m.startedPicker.Selected {
					s := m.startedPicker.Time.Format("2006-01-02")
					startedAt = &s
				}
				if m.finishedPicker.Selected {
					f := m.finishedPicker.Time.Format("2006-01-02")
					finishedAt = &f
				}

				m.confirming = true
				m.pendingPages = pages
				m.pendingStarted = startedAt
				m.pendingFinished = finishedAt
				return m, nil
			}
			if m.focus == focusStarted {
				m.startedPicker.SelectDate()
				return m, nil
			}
			if m.focus == focusFinished {
				m.finishedPicker.SelectDate()
				return m, nil
			}
		}

		switch m.focus {
		case focusPage:
			var cmd tea.Cmd
			m.pageInput, cmd = m.pageInput.Update(msg)
			return m, cmd
		case focusStarted:
			var cmd tea.Cmd
			m.startedPicker, cmd = m.startedPicker.Update(msg)
			return m, cmd
		case focusFinished:
			var cmd tea.Cmd
			m.finishedPicker, cmd = m.finishedPicker.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *Model) updateProgress(pages int, startedAt, finishedAt *string) tea.Cmd {
	client := m.client
	ub := m.userBook
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if ub == nil || len(ub.UserBookReads) == 0 {
			return progressUpdatedMsg{err: fmt.Errorf("no active read found")}
		}
		read := ub.UserBookReads[0]
		err := mutations.UpdateUserBookRead(ctx, client, read.ID, &pages)
		if err != nil {
			return progressUpdatedMsg{err: err}
		}
		if startedAt != nil || finishedAt != nil {
			err = mutations.UpdateUserBookReadDates(ctx, client, read.ID, startedAt, finishedAt)
		}
		return progressUpdatedMsg{err: err}
	}
}

func (m *Model) View() string {
	var b strings.Builder

	title := "Book"
	if m.userBook != nil {
		title = m.userBook.Book.Title
	}
	b.WriteString(common.TitleStyle.Render(fmt.Sprintf("Update Progress - %s", title)))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(common.ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n")
	}

	if m.success {
		b.WriteString(common.PanelStyle.Render(common.SuccessStyle.Render("Progress updated!")))
		b.WriteString("\n\n")
		b.WriteString(common.HelpStyle.Render("esc: back"))
		return common.AppStyle.Render(b.String())
	}

	if m.loading {
		b.WriteString(fmt.Sprintf("  %s Updating...\n", m.spinner.View()))
		return common.AppStyle.Render(b.String())
	}

	if m.confirming {
		var confirm strings.Builder
		confirm.WriteString(common.LabelStyle.Render("Update progress?"))
		confirm.WriteString("\n\n")
		confirm.WriteString(common.ValueStyle.Render(fmt.Sprintf("  Page: %d", m.pendingPages)))
		confirm.WriteString("\n")
		if m.pendingStarted != nil {
			confirm.WriteString(common.ValueStyle.Render(fmt.Sprintf("  Started: %s", *m.pendingStarted)))
			confirm.WriteString("\n")
		}
		if m.pendingFinished != nil {
			confirm.WriteString(common.ValueStyle.Render(fmt.Sprintf("  Finished: %s", *m.pendingFinished)))
			confirm.WriteString("\n")
		}
		confirm.WriteString("\n")
		yKey := lipgloss.NewStyle().Foreground(common.ColorSuccess).Bold(true).Render("y")
		nKey := lipgloss.NewStyle().Foreground(common.ColorDanger).Bold(true).Render("n")
		confirm.WriteString(fmt.Sprintf("  %s: yes  %s: no", yKey, nKey))

		overlayBox := common.PanelActiveStyle.
			Width(40).
			Padding(1, 2).
			Render(confirm.String())

		centered := lipgloss.Place(m.width-4, m.height-4,
			lipgloss.Center, lipgloss.Center,
			overlayBox)
		return common.AppStyle.Render(centered)
	}

	if m.userBook != nil {
		var info strings.Builder
		if m.userBook.Book.Pages != nil {
			info.WriteString(fmt.Sprintf("Total pages: %d\n", *m.userBook.Book.Pages))
		}
		if len(m.userBook.UserBookReads) > 0 {
			read := m.userBook.UserBookReads[0]
			if read.ProgressPages != nil {
				info.WriteString(fmt.Sprintf("Current progress: %d pages\n", *read.ProgressPages))
				if m.userBook.Book.Pages != nil && *m.userBook.Book.Pages > 0 {
					pct := float64(*read.ProgressPages) / float64(*m.userBook.Book.Pages)
					bar := common.RenderBar(pct, 30, fmt.Sprintf("%d%%", int(pct*100)))
					info.WriteString(bar)
				}
			}
		}
		if info.Len() > 0 {
			b.WriteString(common.PanelStyle.Render(info.String()))
			b.WriteString("\n\n")
		}
	}

	pageLabel := "Page number:"
	if m.focus == focusPage {
		pageLabel = "> " + pageLabel
	}
	b.WriteString(common.LabelStyle.Render(pageLabel))
	b.WriteString("\n")
	if m.focus == focusPage {
		b.WriteString(common.FocusedBorderStyle.Render(m.pageInput.View()))
	} else {
		b.WriteString(common.BlurredBorderStyle.Render(m.pageInput.View()))
	}
	b.WriteString("\n\n")

	startedLabel := "Started reading:"
	if m.focus == focusStarted {
		startedLabel = "> " + startedLabel
	}
	var startedPanel strings.Builder
	startedPanel.WriteString(common.LabelStyle.Render(startedLabel))
	if m.startedPicker.Selected {
		startedPanel.WriteString(common.ValueStyle.Render(
			fmt.Sprintf("  [%s]", m.startedPicker.Time.Format("2006-01-02"))))
	}
	startedPanel.WriteString("\n")
	if m.focus == focusStarted {
		startedPanel.WriteString(common.PanelActiveStyle.Render(m.startedPicker.View()))
	} else {
		startedPanel.WriteString(common.PanelStyle.Render(m.startedPicker.View()))
	}

	finishedLabel := "Finished reading:"
	if m.focus == focusFinished {
		finishedLabel = "> " + finishedLabel
	}
	var finishedPanel strings.Builder
	finishedPanel.WriteString(common.LabelStyle.Render(finishedLabel))
	if m.finishedPicker.Selected {
		finishedPanel.WriteString(common.ValueStyle.Render(
			fmt.Sprintf("  [%s]", m.finishedPicker.Time.Format("2006-01-02"))))
	}
	finishedPanel.WriteString("\n")
	if m.focus == focusFinished {
		finishedPanel.WriteString(common.PanelActiveStyle.Render(m.finishedPicker.View()))
	} else {
		finishedPanel.WriteString(common.PanelStyle.Render(m.finishedPicker.View()))
	}

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, startedPanel.String(), "  ", finishedPanel.String()))
	b.WriteString("\n\n")

	b.WriteString(common.HelpStyle.Render("tab: next field | enter: update/select date | esc: back"))

	return common.AppStyle.Render(b.String())
}
