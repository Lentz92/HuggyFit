// internal/tui/view.go

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the entire application UI
func (m Model) View() string {
	if m.quitting {
		return "Thanks for using HuggyFit!\n"
	}

	mainContent := m.renderMainView()

	if m.getHelpVisibility() {
		return m.overlayHelpOnContent(mainContent)
	}

	return mainContent
}

// renderMainView builds the primary UI content
func (m Model) renderMainView() string {
	var s strings.Builder

	s.WriteString(m.renderHeader())
	s.WriteString(m.renderErrorIfPresent())
	s.WriteString(m.renderSearchIfActive())
	s.WriteString(m.renderMainContent())
	s.WriteString(m.renderControlHints())

	return s.String()
}

// renderHeader returns the application header
func (m Model) renderHeader() string {
	return titleStyle.Render("🤗 HuggyFit - GPU Memory Calculator") + "\n\n"
}

// renderErrorIfPresent returns error message if there is one
func (m Model) renderErrorIfPresent() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n"
	}
	return ""
}

// renderSearchIfActive returns the search input if search mode is active
func (m Model) renderSearchIfActive() string {
	if m.searchMode {
		return m.textInput.View() + "\n\n"
	}
	return ""
}

// renderMainContent returns either loading indicator or main panels
func (m Model) renderMainContent() string {
	if m.loading {
		return fmt.Sprintf("%s Loading...", m.spinner.View())
	}

	// Apply dynamic widths based on terminal size
	currentListStyle := listStyle.Width(getListWidth(m.width))
	currentDetailStyle := detailStyle.Width(getDetailWidth(m.width))

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		currentListStyle.Render(m.renderModelList()),
		currentDetailStyle.Render(m.renderModelDetails()),
	)
}

// renderControlHints returns the navigation help text
func (m Model) renderControlHints() string {
	return m.renderControls()
}

// overlayHelpOnContent combines help content with main content
func (m Model) overlayHelpOnContent(mainContent string) string {
	helpContent := m.renderHelp()

	mainLines := strings.Split(mainContent, "\n")
	helpLines := strings.Split(helpContent, "\n")

	dimensions := m.calculateOverlayDimensions(helpLines)
	return m.renderOverlay(mainLines, helpLines, dimensions)
}

// overlayDimensions holds the calculated dimensions for the help overlay
type overlayDimensions struct {
	helpWidth  int
	helpHeight int
	vPad       int
	hPad       int
}

// calculateOverlayDimensions determines the size and position of the help overlay
func (m Model) calculateOverlayDimensions(helpLines []string) overlayDimensions {
	helpHeight := len(helpLines)
	helpWidth := 0
	for _, line := range helpLines {
		if len(line) > helpWidth {
			helpWidth = len(line)
		}
	}

	vPad := max(0, (m.height-helpHeight)/2)
	hPad := max(0, (m.width-helpWidth)/2)

	return overlayDimensions{
		helpWidth:  helpWidth,
		helpHeight: helpHeight,
		vPad:       vPad,
		hPad:       hPad,
	}
}

// renderOverlay combines main content with help overlay
func (m Model) renderOverlay(mainLines, helpLines []string, dims overlayDimensions) string {
	var result strings.Builder

	for i := 0; i < m.height; i++ {
		if i >= dims.vPad && i < dims.vPad+dims.helpHeight {
			result.WriteString(m.renderOverlayLine(
				mainLines,
				helpLines[i-dims.vPad],
				i,
				dims,
			))
		} else if i < len(mainLines) {
			result.WriteString(mainLines[i])
		}
		result.WriteString("\n")
	}

	return result.String()
}

// renderOverlayLine handles rendering a single line of the overlay
func (m Model) renderOverlayLine(mainLines []string, helpLine string, lineNum int, dims overlayDimensions) string {
	var line strings.Builder

	if lineNum < len(mainLines) {
		mainLine := mainLines[lineNum]

		// Write content before help
		if dims.hPad > 0 {
			if len(mainLine) > dims.hPad {
				line.WriteString(mainLine[:dims.hPad])
			} else {
				line.WriteString(mainLine)
				line.WriteString(strings.Repeat(" ", dims.hPad-len(mainLine)))
			}
		}

		// Write help content
		line.WriteString(helpLine)

		// Write remaining main content
		if len(mainLine) > dims.hPad+dims.helpWidth {
			line.WriteString(mainLine[dims.hPad+dims.helpWidth:])
		}
	} else {
		// No main content, just help with padding
		line.WriteString(strings.Repeat(" ", dims.hPad))
		line.WriteString(helpLine)
	}

	return line.String()
}
