package lists

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

func (m *Model) View() string {
	if m.loading {
		return common.AppStyle.Render(
			fmt.Sprintf("\n  %s Loading lists...\n", m.spinner.View()),
		)
	}

	if m.mode == modeCreate {
		var b strings.Builder
		b.WriteString(common.TitleStyle.Render("My Lists"))
		b.WriteString("\n\n")
		var create strings.Builder
		create.WriteString(common.LabelStyle.Render("New list name:"))
		create.WriteString("\n")
		create.WriteString(common.FocusedBorderStyle.Render(m.nameInput.View()))
		create.WriteString("\n\n")
		create.WriteString(common.HelpStyle.Render("enter: create | esc: cancel"))
		b.WriteString(common.RenderActivePanel("New List", create.String(), 0))
		return common.AppStyle.Render(b.String())
	}

	if m.mode == modeAddBook {
		body := m.renderNormalView()

		fg := m.renderAddBookOverlay()
		composed := overlay.Composite(fg, body, overlay.Center, overlay.Center, 0, 0)
		return common.AppStyle.Render(composed)
	}

	if m.mode == modePrivacy {
		body := m.renderNormalView()
		fg := m.renderPrivacyOverlay()
		composed := overlay.Composite(fg, body, overlay.Center, overlay.Center, 0, 0)
		return common.AppStyle.Render(composed)
	}

	if m.mode == modeConfirm {
		body := m.renderNormalView()
		fg := common.RenderConfirmOverlay(m.confirm.Message, m.confirm.Cursor, 50)
		composed := overlay.Composite(fg, body, overlay.Center, overlay.Center, 0, 0)
		return common.AppStyle.Render(composed)
	}

	return common.AppStyle.Render(m.renderNormalView())
}

func (m *Model) renderNormalView() string {
	fbW := m.width - 2
	if fbW < 50 {
		fbW = 80
	}
	fbH := m.height
	if fbH < 10 {
		fbH = 30
	}
	m.flexBox.SetWidth(fbW)
	m.flexBox.SetHeight(fbH)
	m.flexBox.ForceRecalculate()

	leftCell := m.flexBox.GetRow(0).GetCell(0)
	rightCell := m.flexBox.GetRow(0).GetCell(1)
	leftW := leftCell.GetWidth()
	rightW := rightCell.GetWidth()
	cellH := leftCell.GetHeight()

	panelH := cellH
	if panelH < 5 {
		panelH = 5
	}

	listH := panelH - 3
	if listH < 3 {
		listH = 3
	}

	m.list.SetSize(leftW-4, listH)
	m.bookList.SetSize(rightW-4, listH-5)

	listView := m.list.View()

	var leftPanel string
	if m.focusRight {
		leftPanel = common.RenderPanel("My Lists", listView, leftW, panelH)
	} else {
		leftPanel = common.RenderActivePanel("My Lists", listView, leftW, panelH)
	}
	leftCell.SetContent(leftPanel)

	var rightContent string
	var listMeta strings.Builder

	if item, ok := m.list.SelectedItem().(listItem); ok {
		subtext := lipgloss.NewStyle().Foreground(common.ColorSubtext)
		if m.user != nil {
			listMeta.WriteString(subtext.Render("@" + m.user.Username))
			listMeta.WriteString("\n")
		}
		if item.data.Description != nil && *item.data.Description != "" {
			listMeta.WriteString(subtext.Render(*item.data.Description))
			listMeta.WriteString("\n")
		}
		listMeta.WriteString("\n")
	}

	if m.err != nil {
		rightContent = listMeta.String() + common.ErrorStyle.Render("Error: "+m.err.Error())
	} else if m.booksLoading {
		rightContent = listMeta.String() + fmt.Sprintf("  %s Loading...\n", m.spinner.View())
	} else {
		rightContent = listMeta.String() + m.bookList.View()
	}

	listTitle := "Books"
	if item, ok := m.list.SelectedItem().(listItem); ok {
		listTitle = item.data.Name
	}

	var rightPanel string
	if m.focusRight {
		rightPanel = common.RenderActivePanel(listTitle, rightContent, rightW, panelH)
	} else {
		rightPanel = common.RenderPanel(listTitle, rightContent, rightW, panelH)
	}
	rightCell.SetContent(rightPanel)

	return m.flexBox.Render()
}

func (m *Model) renderAddBookOverlay() string {
	listName := "List"
	if item, ok := m.list.SelectedItem().(listItem); ok {
		listName = item.data.Name
	}

	w := m.width - 10
	if w < 50 {
		w = 50
	}
	if w > 90 {
		w = 90
	}

	var content strings.Builder
	content.WriteString(common.LabelStyle.Render("Search:"))
	content.WriteString("\n")
	content.WriteString(common.FocusedBorderStyle.Render(m.searchInput.View()))
	content.WriteString("\n\n")

	if m.addSuccess {
		content.WriteString(lipgloss.NewStyle().Foreground(common.ColorSuccess).Bold(true).Render("Added to list!"))
		content.WriteString("\n\n")
	}
	if m.addErr != nil {
		content.WriteString(common.ErrorStyle.Render("Error: " + m.addErr.Error()))
		content.WriteString("\n\n")
	}

	if m.searching {
		content.WriteString(fmt.Sprintf("  %s Searching...\n", m.spinner.View()))
	} else if len(m.searchResults) > 0 {
		tableW := w - 6
		if tableW < 40 {
			tableW = 40
		}
		m.searchTable.SetColumns(searchTableColumns(tableW))
		m.searchTable.SetWidth(tableW)
		tableH := m.height - 16
		if tableH < 5 {
			tableH = 5
		}
		m.searchTable.SetHeight(tableH)
		content.WriteString(m.searchTable.View())
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(common.HelpStyle.Render("enter: search/add | tab: switch focus | esc: cancel"))

	return common.RenderActivePanel("Add Book to "+listName, content.String(), w)
}

func (m *Model) renderPrivacyOverlay() string {
	w := 40
	privacyOptions := api.AllPrivacySettings()

	listName := "List"
	if item, ok := m.list.SelectedItem().(listItem); ok {
		listName = item.data.Name
	}

	var sel strings.Builder
	for i, p := range privacyOptions {
		cursor := "  "
		style := common.ValueStyle
		if i == m.privacyCursor {
			cursor = lipgloss.NewStyle().Foreground(common.ColorPrimary).Render("> ")
			style = lipgloss.NewStyle().Foreground(common.ColorPrimary).Bold(true)
		}
		sel.WriteString(cursor + style.Render(p.String()) + "\n")
	}
	sel.WriteString("\n")
	sel.WriteString(common.HelpStyle.Render("j/k: navigate | enter: select | esc: cancel"))

	return common.RenderActivePanel("Privacy: "+listName, sel.String(), w)
}

// HelpBindings returns page-specific keybindings for the global help bar.
func (m *Model) HelpBindings() []key.Binding {
	if m.mode == modeAddBook {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "search/add")),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch focus")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		}
	}
	if m.mode == modePrivacy {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		}
	}
	if m.focusRight {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "book details")),
			key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "remove from list")),
		}
	}
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new list")),
		key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add book")),
		key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "privacy")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	}
}
