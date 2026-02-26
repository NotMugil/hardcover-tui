package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/NotMugil/hardcover-tui/internal/common"
)

type ConfirmState struct {
	Active  bool
	Message string
	Action  string
	Cursor  int
}

func NewConfirm(message, action string) ConfirmState {
	return ConfirmState{
		Active:  true,
		Message: message,
		Action:  action,
		Cursor:  1,
	}
}

func (c *ConfirmState) HandleKey(key string) (confirmed bool, handled bool) {
	switch key {
	case "esc":
		c.Active = false
		return false, true
	case "up", "k", "left", "h":
		c.Cursor = 0
		return false, true
	case "down", "j", "right", "l":
		c.Cursor = 1
		return false, true
	case "y":
		c.Active = false
		return true, true
	case "n":
		c.Active = false
		return false, true
	case "enter":
		c.Active = false
		return c.Cursor == 0, true
	}
	return false, true
}

func RenderConfirmOverlay(message string, cursor int, w int) string {
	if w < 30 {
		w = 30
	}
	if w > 50 {
		w = 50
	}

	msgStyle := lipgloss.NewStyle().
		Foreground(common.ColorText).
		Width(w - 6).
		Align(lipgloss.Center)

	yesStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
	noStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)

	if cursor == 0 {
		yesStyle = lipgloss.NewStyle().
			Foreground(common.ColorBackground).
			Background(common.ColorDanger).
			Bold(true).
			Padding(0, 2)
	} else {
		yesStyle = yesStyle.Padding(0, 2)
	}

	if cursor == 1 {
		noStyle = lipgloss.NewStyle().
			Foreground(common.ColorBackground).
			Background(common.ColorSuccess).
			Bold(true).
			Padding(0, 2)
	} else {
		noStyle = noStyle.Padding(0, 2)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		yesStyle.Render("Yes"),
		"  ",
		noStyle.Render("No"),
	)

	buttonsRow := lipgloss.NewStyle().
		Width(w - 6).
		Align(lipgloss.Center).
		Render(buttons)

	content := msgStyle.Render(message) + "\n\n" + buttonsRow + "\n\n" +
		common.HelpStyle.Render("y/n | enter: confirm | esc: cancel")

	return common.RenderActivePanel("Confirm", content, w)
}
