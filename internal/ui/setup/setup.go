package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
	"github.com/NotMugil/hardcover-tui/internal/common"
	"github.com/NotMugil/hardcover-tui/internal/keystore"
)

// SetupCompleteMsg is sent when the user has successfully authenticated.
type SetupCompleteMsg struct {
	Token string
}

type state int

const (
	stateInput state = iota
	stateValidating
	stateError
)

type validateMsg struct {
	err error
}

// Model is the setup screen model.
type Model struct {
	textInput textinput.Model
	spinner   spinner.Model
	help      help.Model
	state     state
	err       error
	token     string
	width     int
	height    int
}

// New creates a new setup screen.
func New() *Model {
	ti := textinput.New()
	ti.Placeholder = "Bearer xxxxxxxxxx..."
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Width = 60
	ti.Cursor.Style = common.CursorStyle
	ti.Focus()

	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)

	return &Model{
		textInput: ti,
		spinner:   s,
		help:      common.NewHelp(),
		state:     stateInput,
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

// SetSize updates the available terminal dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.help.Width = w
}

func (m *Model) InputFocused() bool {
	return m.state == stateInput
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.state == stateInput {
				token := strings.TrimSpace(m.textInput.Value())
				if token == "" {
					return m, nil
				}
				m.token = token
				m.state = stateValidating
				return m, tea.Batch(m.spinner.Tick, m.validateToken(token))
			}
			if m.state == stateError {
				m.state = stateInput
				m.err = nil
				m.textInput.SetValue("")
				m.textInput.Focus()
				return m, textinput.Blink
			}
		}

	case validateMsg:
		if msg.err != nil {
			m.state = stateError
			m.err = msg.err
			return m, nil
		}
		if err := keystore.Save(m.token); err != nil {
			m.state = stateError
			m.err = fmt.Errorf("failed to save key: %w", err)
			return m, nil
		}
		return m, func() tea.Msg {
			return SetupCompleteMsg{Token: m.token}
		}

	case spinner.TickMsg:
		if m.state == stateValidating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	if m.state == stateInput {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) validateToken(token string) tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(token)
		ctx, cancel := makeContext()
		defer cancel()
		_, err := queries.GetMe(ctx, client)
		return validateMsg{err: err}
	}
}

func (m *Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	h := m.height
	if h <= 0 {
		h = 24
	}

	var sections []string

	logo := common.LogoStyle.Render(common.Logo)
	sections = append(sections, logo, "")

	switch m.state {
	case stateInput:
		sections = append(sections,
			common.QuoteStyle.Render("An Unofficial Hardcover TUI client"),
			"",
			common.LabelStyle.Render("Enter your API token:"),
			"",
			common.FocusedBorderStyle.Render(m.textInput.View()),
			"",
			common.ValueStyle.Render("Get your token from https://hardcover.app/account/api"),
			"",
			m.help.ShortHelpView([]key.Binding{
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "continue")),
				key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
			}),
		)

	case stateValidating:
		sections = append(sections,
			fmt.Sprintf("%s Validating token...", m.spinner.View()),
		)

	case stateError:
		sections = append(sections,
			common.ErrorStyle.Render("Authentication failed"),
			"",
		)
		if m.err != nil {
			sections = append(sections,
				common.ValueStyle.Render(m.err.Error()),
				"",
			)
		}
		sections = append(sections,
			m.help.ShortHelpView([]key.Binding{
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "try again")),
				key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
			}),
		)
	}

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}
