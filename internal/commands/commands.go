package commands

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
)

const DefaultTimeout = 30 * time.Second

// Deps holds the shared dependencies that command factories need.
type Deps struct {
	Client *api.Client
	User   *api.User
}

// makeCtx creates a context with the standard timeout.
func makeCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DefaultTimeout)
}

// Run wraps a blocking API call into a tea.Cmd using generics.
func Run[T any](fn func(ctx context.Context) (T, error), wrap func(T, error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := makeCtx()
		defer cancel()
		result, err := fn(ctx)
		return wrap(result, err)
	}
}

// RunVoid wraps a blocking API call with no return value into a tea.Cmd.
func RunVoid(fn func(ctx context.Context) error, wrap func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := makeCtx()
		defer cancel()
		err := fn(ctx)
		return wrap(err)
	}
}
