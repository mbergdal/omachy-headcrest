package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PackageChoice describes a manifest package shown in the install selector.
type PackageChoice struct {
	Name        string
	Label       string
	Description string
	Kind        string
	Installed   bool
	Selected    bool
}

// PackageSelectorModel renders and updates the package checklist.
type PackageSelectorModel struct {
	Items  []PackageChoice
	cursor int
	offset int
}

func NewPackageSelectorModel(items []PackageChoice) PackageSelectorModel {
	return PackageSelectorModel{Items: append([]PackageChoice(nil), items...)}
}

func (m PackageSelectorModel) HasItems() bool {
	return len(m.Items) > 0
}

func (m PackageSelectorModel) SelectedNames() []string {
	selected := []string{}
	for _, item := range m.Items {
		if item.Selected {
			selected = append(selected, item.Name)
		}
	}
	return selected
}

func (m PackageSelectorModel) SelectedCount() int {
	count := 0
	for _, item := range m.Items {
		if item.Selected {
			count++
		}
	}
	return count
}

func (m *PackageSelectorModel) Move(delta, height int) {
	if len(m.Items) == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.Items) {
		m.cursor = len(m.Items) - 1
	}
	m.ensureVisible(visibleSelectorRows(height))
}

func (m *PackageSelectorModel) Toggle() {
	if len(m.Items) == 0 {
		return
	}
	m.Items[m.cursor].Selected = !m.Items[m.cursor].Selected
}

func (m *PackageSelectorModel) SelectAll(selected bool) {
	for i := range m.Items {
		m.Items[i].Selected = selected
	}
}

func (m *PackageSelectorModel) Invert() {
	for i := range m.Items {
		m.Items[i].Selected = !m.Items[i].Selected
	}
}

func (m *PackageSelectorModel) View(width, height int) string {
	if width < 40 {
		width = 40
	}
	rows := visibleSelectorRows(height)
	m.ensureVisible(rows)

	var b strings.Builder
	b.WriteString(splashLogo.Render(logo))
	b.WriteString("\n")
	b.WriteString(splashSubtitle.Render("  Select packages to install and configure"))
	b.WriteString("\n\n")
	b.WriteString(splashSection.Render("  Missing packages are selected by default. Installed packages are unchecked."))
	b.WriteString("\n\n")

	end := m.offset + rows
	if end > len(m.Items) {
		end = len(m.Items)
	}
	for i := m.offset; i < end; i++ {
		item := m.Items[i]
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		check := "[ ]"
		if item.Selected {
			check = "[x]"
		}
		status := "missing"
		statusStyle := lipgloss.NewStyle().Foreground(colorWarning)
		if item.Installed {
			status = "installed"
			statusStyle = lipgloss.NewStyle().Foreground(colorMuted)
		}

		label := lipgloss.NewStyle().Foreground(colorText).Render(item.Label)
		detail := splashItem.Render(item.Description)
		if item.Description == "" {
			detail = splashItem.Render(item.Kind)
		}

		b.WriteString(fmt.Sprintf("  %s %s %-26s %-12s %s\n",
			cursor,
			check,
			label,
			statusStyle.Render(status),
			detail,
		))
	}
	if m.offset > 0 || end < len(m.Items) {
		b.WriteString(splashItem.Render(fmt.Sprintf("\n  Showing %d-%d of %d", m.offset+1, end, len(m.Items))))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(splashSection.Render(fmt.Sprintf("  Selected %d of %d packages", m.SelectedCount(), len(m.Items))))
	b.WriteString("\n")
	b.WriteString(splashPrompt.Render("  [space] toggle  [a] all  [n] none  [i] invert  [Enter] install  [q] quit"))
	b.WriteString("\n")
	return b.String()
}

func visibleSelectorRows(height int) int {
	rows := height - 13
	if rows < 8 {
		rows = 8
	}
	return rows
}

func (m *PackageSelectorModel) ensureVisible(rows int) {
	if rows < 1 {
		rows = 1
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+rows {
		m.offset = m.cursor - rows + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}
