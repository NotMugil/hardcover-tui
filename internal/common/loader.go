package common

import (
	"math"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

const (
	loaderFPS       = 60
	loaderFrequency = 5.0
	loaderDamping   = 0.4
)

type LoaderFrameMsg time.Time

var loaderQuotes = []string{
	"A reader lives a thousand lives before he dies...",
	"So many books, so little time.",
	"Books are a uniquely portable magic.",
	"Reading is dreaming with open eyes.",
	"One more chapter...",
	"There is no friend as loyal as a book.",
	"A book is a dream you hold in your hands.",
	"The world was hers for the reading.",
	"Books are mirrors: you see yourself in them.",
	"I have always imagined paradise as a library.",
	"We read to know we are not alone.",
	"Between the pages of a book is a lovely place.",
	"Reading gives us someplace to go when we have to stay.",
	"Today a reader, tomorrow a leader.",
	"Books are the quietest friends.",
	"Lost in a good book...",
	"Turning pages, turning worlds.",
	"Let the story unfold...",
}

type Loader struct {
	spring harmonica.Spring
	posX   float64
	velX   float64
	frame  int
	active bool
	quote  string
}

func NewLoader() Loader {
	return Loader{
		spring: harmonica.NewSpring(harmonica.FPS(loaderFPS), loaderFrequency, loaderDamping),
		active: false,
	}
}

func (l *Loader) Start() tea.Cmd {
	l.active = true
	l.frame = 0
	l.posX = 0
	l.velX = 0
	l.quote = loaderQuotes[rand.Intn(len(loaderQuotes))]
	return l.tick()
}

func (l *Loader) Stop() {
	l.active = false
}

func (l *Loader) Active() bool {
	return l.active
}

func (l *Loader) tick() tea.Cmd {
	return tea.Tick(time.Second/loaderFPS, func(t time.Time) tea.Msg {
		return LoaderFrameMsg(t)
	})
}

func (l *Loader) Update() tea.Cmd {
	if !l.active {
		return nil
	}

	targetX := 1.0
	if (l.frame/30)%2 == 0 {
		targetX = 0.0
	}
	l.posX, l.velX = l.spring.Update(l.posX, l.velX, targetX)

	l.frame++
	return l.tick()
}

func (l *Loader) View(width, height int) string {
	if width < 10 {
		width = 80
	}
	if height < 5 {
		height = 24
	}

	sprites := []string{"█", "▓", "▒", "░"}
	idx := l.frame / 8 % len(sprites)

	rangeW := float64(width) * 0.4
	offsetX := float64(width)*0.3 + l.posX*rangeW

	col := int(math.Round(offsetX))
	if col < 0 {
		col = 0
	}
	spriteLen := 5
	if col > width-spriteLen-1 {
		col = width - spriteLen - 1
	}

	spriteChar := sprites[idx]
	sprite := strings.Repeat(spriteChar, spriteLen)

	barStyle := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	spriteLine := strings.Repeat(" ", col) + barStyle.Render(sprite)

	quoteText := l.quote
	if len(quoteText) > width-4 {
		quoteText = quoteText[:width-7] + "..."
	}

	centeredQuote := lipgloss.PlaceHorizontal(width, lipgloss.Center, QuoteStyle.Render(quoteText))

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		spriteLine,
		"",
		centeredQuote,
	)

	contentH := lipgloss.Height(content)
	topPad := (height - contentH) / 2
	if topPad < 0 {
		topPad = 0
	}

	padded := strings.Repeat("\n", topPad) + content
	return lipgloss.NewStyle().Width(width).Height(height).Render(padded)
}
