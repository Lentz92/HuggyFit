// internal/tui/list_view.go

package tui

import (
	"fmt"
	"strings"
)

// renderModelList renders the list of available models
func (m Model) renderModelList() string {
	var s strings.Builder

	// Handle empty list case
	if len(m.modelIDs) == 0 {
		return listStyle.Render("No models found")
	}

	// Get current page bounds
	start, end := getPageBounds(m.cursor, len(m.modelIDs))

	// Calculate available width for content
	// Account for: borders (2), padding (2), cursor indicator (2), margin (1)
	contentWidth := getListWidth(m.width) - 7

	// Build header
	s.WriteString("Available Models\n")
	s.WriteString(strings.Repeat("─", contentWidth))
	s.WriteString("\n")

	// Build model list
	for i := start; i < end; i++ {
		s.WriteString(m.renderModelListItem(i, contentWidth))
	}

	// Build footer
	s.WriteString(m.renderModelListFooter(start, contentWidth))

	return listStyle.Render(s.String())
}

// renderModelListItem renders a single model list item
func (m Model) renderModelListItem(index int, width int) string {
	// Cursor indicator
	cursor := " "
	if m.cursor == index {
		cursor = ">"
	}

	// Format model ID with smart truncation
	modelID := m.modelIDs[index]
	maxIDWidth := width - 2 // Account for cursor and space

	// Smart truncation that preserves important parts
	if len(modelID) > maxIDWidth {
		// Try to preserve organization/model structure
		parts := strings.Split(modelID, "/")
		if len(parts) > 1 {
			// Keep first and last parts if possible
			first := parts[0]
			last := parts[len(parts)-1]
			if len(first)+len(last)+4 <= maxIDWidth {
				modelID = first + "/../" + last
			} else {
				modelID = modelID[:maxIDWidth-3] + "..."
			}
		} else {
			modelID = modelID[:maxIDWidth-3] + "..."
		}
	}
	paddedModel := fmt.Sprintf("%-*s", maxIDWidth, modelID)

	// Apply styling
	if m.cursor == index {
		return fmt.Sprintf("%s %s\n", cursor, selectedStyle.Render(paddedModel))
	}
	return fmt.Sprintf("%s %s\n", cursor, paddedModel)
}

// renderModelListFooter renders the pagination footer
func (m Model) renderModelListFooter(startIndex int, width int) string {
	var s strings.Builder

	// Add separator
	s.WriteString(strings.Repeat("─", width))
	s.WriteString("\n")

	// Calculate pagination
	currentPage := (startIndex / itemsPerPage) + 1
	totalPages := (len(m.modelIDs) + itemsPerPage - 1) / itemsPerPage

	// Add page information
	pageInfo := fmt.Sprintf("Page %d of %d (%d models)",
		currentPage,
		totalPages,
		len(m.modelIDs))

	s.WriteString(pageInfo)

	return s.String()
}
