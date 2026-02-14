package common

import (
	_ "embed"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	ColorPrimary    = lipgloss.Color("#6366f1")
	ColorSecondary  = lipgloss.Color("#7C3AED")
	ColorAccent     = lipgloss.Color("#F59E0B")
	ColorSuccess    = lipgloss.Color("#10B981")
	ColorWarning    = lipgloss.Color("#F59E0B")
	ColorDanger     = lipgloss.Color("#EF4444")
	ColorMuted      = lipgloss.Color("#4B5563")
	ColorText       = lipgloss.Color("#E5E7EB")
	ColorSubtext    = lipgloss.Color("#9CA3AF")
	ColorBorder     = lipgloss.Color("#374151")
	ColorBackground = lipgloss.Color("#0D1117")
	ColorSurface    = lipgloss.Color("#161B22")
	ColorHighlight  = lipgloss.Color("#1C2333")

	ColorWantToRead       = lipgloss.Color("#3B82F6")
	ColorCurrentlyReading = lipgloss.Color("#10B981")
	ColorRead             = lipgloss.Color("#8B5CF6")
	ColorPaused           = lipgloss.Color("#F59E0B")
	ColorDNF              = lipgloss.Color("#EF4444")
	ColorIgnored          = lipgloss.Color("#4B5563")
)

// Set Status color based on status ID
func StatusColor(statusID int) lipgloss.Color {
	switch statusID {
	case 1:
		return ColorWantToRead
	case 2:
		return ColorCurrentlyReading
	case 3:
		return ColorRead
	case 4:
		return ColorPaused
	case 5:
		return ColorDNF
	case 6:
		return ColorIgnored
	default:
		return ColorMuted
	}
}

// Layouts and borders
var (
	AppStyle = lipgloss.NewStyle().
			Padding(0, 1)

	BasePanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 1)

	PanelStyle       = BasePanelStyle.BorderForeground(ColorBorder)
	PanelActiveStyle = BasePanelStyle.BorderForeground(ColorPrimary)

	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBackground).
			Background(ColorPrimary).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Background(ColorSurface).
			Padding(0, 1)

	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	BlurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorBorder).
				Padding(0, 1)

	CursorStyle = lipgloss.NewStyle().Foreground(ColorPrimary)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	ProgressFilledStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorBorder)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)
)

// Typography
var (
	BoldTextStyle = lipgloss.NewStyle().Bold(true)
	TitleStyle    = BoldTextStyle.Foreground(ColorPrimary)
	LabelStyle    = BoldTextStyle.Foreground(ColorText)
	SubtitleStyle = lipgloss.NewStyle().Foreground(ColorSubtext)
	ValueStyle    = lipgloss.NewStyle().Foreground(ColorSubtext)
	HelpStyle     = lipgloss.NewStyle().Foreground(ColorMuted)
	QuoteStyle    = HelpStyle.Italic(true)
	SuccessStyle  = BoldTextStyle.Foreground(ColorSuccess)
	ErrorStyle    = BoldTextStyle.Foreground(ColorDanger)
)

// ASCII Logo at Setup
var (
	//go:embed banner.txt
	Logo      string
	LogoStyle = TitleStyle
)

// Help or Keybindings
func HelpStyles() help.Styles {
	return help.Styles{
		ShortKey:       lipgloss.NewStyle().Foreground(ColorPrimary),
		ShortDesc:      lipgloss.NewStyle().Foreground(ColorMuted),
		ShortSeparator: lipgloss.NewStyle().Foreground(ColorBorder),
		FullKey:        lipgloss.NewStyle().Foreground(ColorPrimary),
		FullDesc:       lipgloss.NewStyle().Foreground(ColorSubtext),
		FullSeparator:  lipgloss.NewStyle().Foreground(ColorBorder),
		Ellipsis:       lipgloss.NewStyle().Foreground(ColorMuted),
	}
}

func NewHelp() help.Model {
	h := help.New()
	h.ShortSeparator = " • "
	h.FullSeparator = "    "
	h.Styles = HelpStyles()
	return h
}
