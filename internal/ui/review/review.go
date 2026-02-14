package review

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/mutations"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

type reviewSavedMsg struct {
	err error
}

// Model is the review screen model.
type Model struct {
	client   *api.Client
	user     *api.User
	userBook *api.UserBook
	textarea textarea.Model
	spinner  spinner.Model
	loading  bool
	err      error
	success  bool
	editing  bool // true when textarea is focused
}

// New creates a new review screen.
func New(client *api.Client, user *api.User, ub *api.UserBook) *Model {
	ta := textarea.New()
	ta.Placeholder = "Write your review..."
	ta.SetWidth(60)
	ta.SetHeight(10)
	ta.Cursor.Style = common.CursorStyle
	ta.Focus()

	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)

	m := &Model{
		client:   client,
		user:     user,
		userBook: ub,
		textarea: ta,
		spinner:  s,
		editing:  true,
	}

	if ub != nil && ub.Review != nil {
		ta.SetValue(*ub.Review)
		m.textarea = ta
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}

// SetSize updates the available terminal dimensions.
func (m *Model) SetSize(w, h int) {
	if w > 8 {
		m.textarea.SetWidth(w - 8)
	}
}

// InputFocused returns true when the textarea is actively being edited.
func (m *Model) InputFocused() bool {
	return m.editing && !m.loading && !m.success
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case reviewSavedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		m.success = true
		return m, common.NotifyCmd(common.NotifySuccess, "Review saved")

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
			return m, nil
		}

		switch msg.String() {
		case "ctrl+s":
			review := strings.TrimSpace(m.textarea.Value())
			if review == "" {
				return m, nil
			}
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.saveReview(review))
		case "esc":
			if m.editing {
				m.editing = false
				m.textarea.Blur()
				return m, nil
			}
			return m, nil
		case "i", "enter":
			if !m.editing {
				m.editing = true
				m.textarea.Focus()
				return m, textarea.Blink
			}
		}

		if m.editing {
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

		return m, nil
	}
	return m, nil
}

func (m *Model) saveReview(review string) tea.Cmd {
	client := m.client
	ub := m.userBook
	return func() tea.Msg {
		if ub == nil {
			return reviewSavedMsg{err: fmt.Errorf("no book to review")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.UpdateUserBookReview(ctx, client, ub.ID, review, false)
		return reviewSavedMsg{err: err}
	}
}

func (m *Model) View() string {
	var b strings.Builder

	title := "Book"
	if m.userBook != nil {
		title = m.userBook.Book.Title
	}
	b.WriteString(common.TitleStyle.Render(fmt.Sprintf("Review - %s", title)))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(common.ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n")
	}

	if m.success {
		b.WriteString(common.PanelStyle.Render(common.SuccessStyle.Render("Review saved!")))
		b.WriteString("\n\n")
		b.WriteString(common.HelpStyle.Render("esc: back"))
		return common.AppStyle.Render(b.String())
	}

	if m.loading {
		b.WriteString(fmt.Sprintf("  %s Saving...\n", m.spinner.View()))
		return common.AppStyle.Render(b.String())
	}

	if m.userBook != nil && m.userBook.Rating != nil {
		b.WriteString(common.LabelStyle.Render("Rating: ") + common.RenderRatingBar(*m.userBook.Rating, 15))
		b.WriteString("\n\n")
	}

	b.WriteString(m.textarea.View())
	b.WriteString("\n\n")
	if m.editing {
		b.WriteString(common.HelpStyle.Render("ctrl+s: save | esc: stop editing"))
	} else {
		b.WriteString(common.HelpStyle.Render("i/enter: edit | ctrl+s: save | esc: back"))
	}

	return common.AppStyle.Render(b.String())
}
