package common

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderBar renders a block-style progress bar: ████░░░░ + label.
// pct is 0.0–1.0, width is the bar character count, label is appended after the bar.
func RenderBar(pct float64, width int, label string) string {
	if width <= 0 {
		width = 20
	}
	if pct > 1 {
		pct = 1
	}
	if pct < 0 {
		pct = 0
	}
	filled := int(pct * float64(width))
	empty := width - filled

	bar := strings.Repeat(ProgressFilledStyle.Render("\u2588"), filled) +
		strings.Repeat(ProgressEmptyStyle.Render("\u2591"), empty)
	if label != "" {
		bar += " " + ProgressFilledStyle.Render(label)
	}
	return bar
}

// RenderRatingBar renders a rating (0–5) as a block bar with "X.Y/5" label.
func RenderRatingBar(rating float64, width int) string {
	label := fmt.Sprintf("%.1f/5", rating)
	return RenderBar(rating/5.0, width, label)
}

// panelOpts configures a bordered panel.
type panelOpts struct {
	borderColor lipgloss.Color
	titleColor  lipgloss.Color
}

// Bordered panels with inline title
func RenderPanel(title, content string, width int, heights ...int) string {
	return renderPanel(title, content, width, firstOr(heights, 0), panelOpts{
		borderColor: ColorBorder,
		titleColor:  ColorPrimary,
	})
}

func RenderActivePanel(title, content string, width int, heights ...int) string {
	return renderPanel(title, content, width, firstOr(heights, 0), panelOpts{
		borderColor: ColorPrimary,
		titleColor:  ColorPrimary,
	})
}

func RenderDimPanel(title, content string, width int, heights ...int) string {
	return renderPanel(title, content, width, firstOr(heights, 0), panelOpts{
		borderColor: lipgloss.Color("#1F2937"),
		titleColor:  ColorMuted,
	})
}

// firstOr returns the first element of s, or fallback if empty.
func firstOr(s []int, fallback int) int {
	if len(s) > 0 {
		return s[0]
	}
	return fallback
}

// clampMin returns max(v, min).
func clampMin(v, min int) int {
	if v < min {
		return min
	}
	return v
}

// repeatStyled renders ch repeated n times with style applied once.
func repeatStyled(ch string, n int, style lipgloss.Style) string {
	if n <= 0 {
		return ""
	}
	return style.Render(strings.Repeat(ch, n))
}

// renderPanel is the single internal implementation for all bordered panels.
func renderPanel(title, content string, width, height int, opts panelOpts) string {
	if width <= 0 {
		width = lipgloss.Width(content) + 4
	}
	innerW := clampMin(width-4, 1) // border(1+1) + padding(1+1)

	wrapped := lipgloss.NewStyle().Width(innerW).Render(content)
	border := lipgloss.NormalBorder()
	bdr := lipgloss.NewStyle().Foreground(opts.borderColor)

	if title == "" {
		style := PanelStyle.Width(width).BorderForeground(opts.borderColor)
		if height > 0 {
			style = style.Height(clampMin(height-2, 1))
		}
		return style.Render(wrapped)
	}

	titleStr := lipgloss.NewStyle().Bold(true).Foreground(opts.titleColor).Render(title)
	pad := clampMin(width-2-lipgloss.Width(titleStr)-4, 0)
	top := bdr.Render(border.TopLeft+border.Top+border.Top+" ") +
		titleStr +
		bdr.Render(" ") +
		repeatStyled(border.Top, pad, bdr) +
		bdr.Render(border.TopRight)

	bottom := bdr.Render(border.BottomLeft) +
		repeatStyled(border.Bottom, width-2, bdr) +
		bdr.Render(border.BottomRight)

	lines := strings.Split(wrapped, "\n")
	if height > 0 {
		innerH := clampMin(height-2, 1)
		for len(lines) < innerH {
			lines = append(lines, "")
		}
		if len(lines) > innerH {
			lines = lines[:innerH]
		}
	}

	var mid strings.Builder
	left := bdr.Render(border.Left)
	right := bdr.Render(border.Right)
	for _, line := range lines {
		gap := clampMin(innerW-lipgloss.Width(line), 0)
		mid.WriteString(left + " " + line + strings.Repeat(" ", gap) + " " + right + "\n")
	}

	return top + "\n" + mid.String() + bottom
}

// Truncate and add ellipsis if string exceeds max width
func Truncate(s string, maxWidth int) string {
	if maxWidth <= 3 {
		return s
	}
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	runes := []rune(s)
	for i := len(runes); i > 0; i-- {
		candidate := string(runes[:i]) + "..."
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}
	return "..."
}
