package common

import tea "github.com/charmbracelet/bubbletea"

// NotifyLevel represents the severity of a notification.
type NotifyLevel string

const (
	NotifyInfo    NotifyLevel = "Info"
	NotifySuccess NotifyLevel = "Success"
	NotifyWarning NotifyLevel = "Warn"
	NotifyError   NotifyLevel = "Error"
)

// NotifyMsg is a message that triggers a toast notification in the app.
// Child screens return this as a tea.Cmd to signal success/error/info.
type NotifyMsg struct {
	Level   NotifyLevel
	Message string
}

// NotifyCmd creates a tea.Cmd that produces a NotifyMsg.
func NotifyCmd(level NotifyLevel, message string) tea.Cmd {
	return func() tea.Msg {
		return NotifyMsg{Level: level, Message: message}
	}
}
