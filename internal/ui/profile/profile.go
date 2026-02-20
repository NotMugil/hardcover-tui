package profile

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/mutations"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
	"github.com/NotMugil/hardcover-tui/internal/common"
	"github.com/NotMugil/hardcover-tui/internal/keystore"
)

type profileLoadedMsg struct {
	user      *api.User
	stats     *api.UserBookAggregate
	avatarArt string
	err       error
}

type profileUpdatedMsg struct {
	err error
}

type loggedOutMsg struct{}

type inputMode int

const (
	modeView inputMode = iota
	modeEdit
)

// Model is the profile screen model.
type Model struct {
	client    *api.Client
	user      *api.User
	stats     *api.UserBookAggregate
	avatarArt string
	spinner   spinner.Model
	loading   bool
	err       error
	mode      inputMode
	nameInput textinput.Model
	bioInput  textinput.Model
	locInput  textinput.Model
	editField int // 0=name, 1=bio, 2=location
}

// New creates a new profile screen.
func New(client *api.Client, user *api.User) *Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)

	ni := textinput.New()
	ni.Placeholder = "Display name"
	ni.Width = 40
	ni.Cursor.Style = common.CursorStyle

	bi := textinput.New()
	bi.Placeholder = "Bio"
	bi.Width = 60
	bi.Cursor.Style = common.CursorStyle

	li := textinput.New()
	li.Placeholder = "Location"
	li.Width = 40
	li.Cursor.Style = common.CursorStyle

	return &Model{
		client:    client,
		user:      user,
		spinner:   s,
		loading:   true,
		nameInput: ni,
		bioInput:  bi,
		locInput:  li,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadProfile())
}

// SetSize updates the available terminal dimensions.
func (m *Model) SetSize(w, h int) {}

// InputFocused returns true when editing profile fields.
func (m *Model) InputFocused() bool {
	return m.mode == modeEdit
}

func (m *Model) loadProfile() tea.Cmd {
	client := m.client
	user := m.user
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		u, err := queries.GetMe(ctx, client)
		if err != nil {
			return profileLoadedMsg{err: err}
		}
		stats, _ := queries.GetUserBookStats(ctx, client, user.ID)

		var avatarArt string
		if u.ImageURL() != "" {
			art, artErr := common.RenderImage(u.ImageURL(), 30, 15)
			if artErr == nil {
				avatarArt = art
			}
		}

		return profileLoadedMsg{user: u, stats: stats, avatarArt: avatarArt}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case profileLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.user = msg.user
		m.stats = msg.stats
		m.avatarArt = msg.avatarArt
		return m, nil

	case profileUpdatedMsg:
		m.loading = false
		m.mode = modeView
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		return m, m.loadProfile()

	case loggedOutMsg:
		return m, tea.Quit

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		if m.mode == modeEdit {
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(m.nameInput.Value())
				bio := strings.TrimSpace(m.bioInput.Value())
				loc := strings.TrimSpace(m.locInput.Value())
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, m.updateProfile(name, bio, loc))
			case "esc":
				m.mode = modeView
				return m, nil
			case "tab":
				m.editField = (m.editField + 1) % 3
				m.nameInput.Blur()
				m.bioInput.Blur()
				m.locInput.Blur()
				switch m.editField {
				case 0:
					m.nameInput.Focus()
				case 1:
					m.bioInput.Focus()
				case 2:
					m.locInput.Focus()
				}
				return m, textinput.Blink
			}
			var cmd tea.Cmd
			switch m.editField {
			case 0:
				m.nameInput, cmd = m.nameInput.Update(msg)
			case 1:
				m.bioInput, cmd = m.bioInput.Update(msg)
			case 2:
				m.locInput, cmd = m.locInput.Update(msg)
			}
			return m, cmd
		}

		if m.loading {
			return m, nil
		}

		switch strings.ToLower(msg.String()) {
		case "e":
			m.mode = modeEdit
			m.editField = 0
			if m.user.Name != nil {
				m.nameInput.SetValue(*m.user.Name)
			}
			if m.user.Bio != nil {
				m.bioInput.SetValue(*m.user.Bio)
			}
			if m.user.Location != nil {
				m.locInput.SetValue(*m.user.Location)
			}
			m.nameInput.Focus()
			return m, textinput.Blink
		case "x":
			return m, m.logout()
		}
	}
	return m, nil
}

func (m *Model) updateProfile(name, bio, location string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.UpdateUserProfile(ctx, client, name, bio, location)
		return profileUpdatedMsg{err: err}
	}
}

func (m *Model) logout() tea.Cmd {
	return func() tea.Msg {
		_ = keystore.Delete()
		return loggedOutMsg{}
	}
}

func (m *Model) View() string {
	if m.loading {
		return common.AppStyle.Render(
			fmt.Sprintf("\n  %s Loading profile...\n", m.spinner.View()),
		)
	}

	var b strings.Builder

	if m.err != nil {
		b.WriteString(common.ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n")
	}

	if m.mode == modeEdit {
		var edit strings.Builder
		edit.WriteString(common.TitleStyle.Render("Edit Profile"))
		edit.WriteString("\n\n")
		edit.WriteString("Name:     ")
		if m.editField == 0 {
			edit.WriteString(common.FocusedBorderStyle.Render(m.nameInput.View()))
		} else {
			edit.WriteString(common.BlurredBorderStyle.Render(m.nameInput.View()))
		}
		edit.WriteString("\nBio:      ")
		if m.editField == 1 {
			edit.WriteString(common.FocusedBorderStyle.Render(m.bioInput.View()))
		} else {
			edit.WriteString(common.BlurredBorderStyle.Render(m.bioInput.View()))
		}
		edit.WriteString("\nLocation: ")
		if m.editField == 2 {
			edit.WriteString(common.FocusedBorderStyle.Render(m.locInput.View()))
		} else {
			edit.WriteString(common.BlurredBorderStyle.Render(m.locInput.View()))
		}
		edit.WriteString("\n\n")
		edit.WriteString(common.HelpStyle.Render("enter: save | tab: next field | esc: cancel"))

		b.WriteString(common.PanelStyle.Render(edit.String()))
		return common.AppStyle.Render(b.String())
	}

	if m.user != nil {
		u := m.user

		var header strings.Builder
		header.WriteString(common.TitleStyle.Render(fmt.Sprintf("@%s", u.Username)))
		if u.Name != nil && *u.Name != "" {
			header.WriteString("  " + common.LabelStyle.Render(*u.Name))
		}
		if u.Pro {
			header.WriteString("  " + common.SuccessStyle.Render("PRO"))
		}
		header.WriteString("\n")
		if u.Flair != nil && *u.Flair != "" {
			header.WriteString(common.ValueStyle.Render(*u.Flair))
		}
		b.WriteString(common.PanelStyle.Render(header.String()))
		b.WriteString("\n")

		if m.avatarArt != "" {
			b.WriteString(common.PanelStyle.Render(m.avatarArt))
			b.WriteString("\n")
		}

		var details strings.Builder
		if u.Bio != nil && *u.Bio != "" {
			details.WriteString(common.ValueStyle.Render(*u.Bio) + "\n\n")
		}
		if u.Location != nil && *u.Location != "" {
			details.WriteString(common.LabelStyle.Render("Location: ") + *u.Location + "\n")
		}
		if u.Link != nil && *u.Link != "" {
			details.WriteString(common.LabelStyle.Render("Link:     ") + *u.Link + "\n")
		}
		details.WriteString(common.LabelStyle.Render("Joined:   ") + u.CreatedAt.Format("January 2006") + "\n")
		details.WriteString(common.LabelStyle.Render("Pronouns: ") + fmt.Sprintf("%s/%s", u.PronounPersonal, u.PronounPossessive))
		b.WriteString(common.PanelStyle.Render(details.String()))
		b.WriteString("\n")

		var stats strings.Builder
		stats.WriteString(common.TitleStyle.Render("Stats") + "\n")
		stats.WriteString(fmt.Sprintf("Books:     %d\n", u.BooksCount))
		stats.WriteString(fmt.Sprintf("Followers: %d\n", u.FollowersCount))
		stats.WriteString(fmt.Sprintf("Following: %d", u.FollowedUsersCount))
		if m.stats != nil {
			if m.stats.Aggregate.Avg.Rating != nil {
				stats.WriteString(fmt.Sprintf("\nAvg Rating: %.1f", *m.stats.Aggregate.Avg.Rating))
			}
		}
		b.WriteString(common.PanelStyle.Render(stats.String()))
	}

	b.WriteString("\n")
	b.WriteString(common.HelpStyle.Render("e: edit profile | x: logout"))

	return common.AppStyle.Render(b.String())
}
