// internal/tui/help_view.go

package tui

import (
	"fmt"
	"strings"
)

// helpSection represents a group of related commands
type helpSection struct {
	category string
	items    []helpItem
}

// helpItem represents a single command and its description
type helpItem struct {
	key  string
	desc string
}

// Help documentation sections
var helpContent = []helpSection{
	{
		category: "Navigation",
		items: []helpItem{
			{"↑/↓, j/k", "Navigate through models"},
			{"PgUp/PgDn", "Jump 10 items"},
			{"Home/End", "Jump to top/bottom"},
			{"Enter", "Select model"},
			{"/", "Search models"},
			{"Esc", "Exit search"},
			{"Tab", "Switch view"},
			{"q", "Quit application"},
		},
	},
	{
		category: "Configuration",
		items: []helpItem{
			{"+/-", "Adjust user count"},
			{"c", "Cycle context length"},
		},
	},
	{
		category: "Display",
		items: []helpItem{
			{"?", "Toggle help"},
		},
	},
}

func (m Model) getHelpDimensions() (width, height int) {
	width = min(60, m.width-4)   // Leave space for borders
	height = min(20, m.height-4) // Leave space for borders
	return
}

// renderHelp renders the help documentation
func (m Model) renderHelp() string {
	if !m.getHelpVisibility() {
		return ""
	}

	width, height := m.getHelpDimensions()

	var s strings.Builder
	s.WriteString("Keyboard Shortcuts\n")
	s.WriteString(strings.Repeat("─", width-4) + "\n\n")

	for _, section := range helpContent {
		s.WriteString(m.renderHelpSection(section))
	}

	s.WriteString("\nPress ? to close help")

	return overlayStyle.
		Width(width).
		Height(height).
		Render(s.String())
}

// renderHelpSection renders a single help section
func (m Model) renderHelpSection(section helpSection) string {
	var s strings.Builder

	// Section header
	s.WriteString(fmt.Sprintf("\n%s:\n", section.category))

	// Section items
	for _, item := range section.items {
		s.WriteString(m.renderHelpItem(item))
	}

	return s.String()
}

// renderHelpItem renders a single help item
func (m Model) renderHelpItem(item helpItem) string {
	// Format: "  key        : description"
	return fmt.Sprintf("  %-12s: %s\n",
		selectedStyle.Render(item.key),
		item.desc)
}

// renderControls renders the compact control hints
func (m Model) renderControls() string {
	var s strings.Builder

	// Navigation controls
	s.WriteString("\nNavigation:\n")
	s.WriteString(helpStyle.Render(
		"↑/↓ or j/k: Navigate • Enter: Select • /: Search • Tab: Switch view • ?: Help • q: Quit\n"))

	// Show configuration controls only when a model is selected
	if m.isModelSelected() {
		s.WriteString(m.renderConfigurationOptions())
	}

	return s.String()
}

// getHelpVisibility returns whether help should be shown
func (m Model) getHelpVisibility() bool {
	return m.showHelp
}

// toggleHelp toggles the help visibility
func (m *Model) toggleHelp() {
	m.showHelp = !m.showHelp
}
