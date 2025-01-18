// internal/tui/details_view.go

package tui

import (
	"fmt"
	"strings"

	"github.com/Lentz92/huggyfit/internal/calculator"
)

func (m Model) renderModelDetails() string {
	if !m.isModelSelected() {
		return detailStyle.Render("Select a model to view details")
	}

	var s strings.Builder

	// Render tabs
	s.WriteString(m.renderTabs())
	s.WriteString("\n\n")

	// Render content based on active tab
	if m.activeTab == 0 {
		s.WriteString(m.renderMemoryDetails())
	} else {
		s.WriteString(m.renderModelInfo())
	}

	return detailStyle.Render(s.String())
}

func (m Model) renderTabs() string {
	tabs := []string{"Memory Requirements", "Model Details"}
	var parts []string

	for i, tab := range tabs {
		if i == m.activeTab {
			parts = append(parts, activeTabStyle.Render("("+tab+")"))
		} else {
			parts = append(parts, tabStyle.Render("("+tab+")"))
		}
	}

	return strings.Join(parts, " ")
}

func (m Model) renderMemoryDetails() string {
	var s strings.Builder

	// Model identification
	s.WriteString("Model: " + m.modelInfo.ModelID + "\n")

	// Configuration section
	s.WriteString("\nConfiguration:\n")
	s.WriteString("- Users: " + valueStyle.Render(fmt.Sprint(m.users)) + "\n")
	s.WriteString("- Context Length: " + valueStyle.Render(formatContextLength(m.contextLen)) + "\n")

	// Memory calculations for each data type
	for _, dtype := range dataTypes {
		s.WriteString(m.renderMemoryCalculation(dtype))
	}

	return s.String()
}

func (m Model) renderMemoryCalculation(dtype calculator.DataType) string {
	var s strings.Builder

	baseMemory, _ := calculator.CalculateGPUMemory(m.modelInfo.ParametersB, dtype)
	kvMemory := m.calculateKVCache(dtype)
	totalMemory := baseMemory + kvMemory

	s.WriteString("\n" + string(dtype) + ":\n")
	s.WriteString("  Base: " + valueStyle.Render(fmt.Sprintf("%.2f GB", baseMemory)) + "\n")
	s.WriteString("  KV Cache: " + valueStyle.Render(fmt.Sprintf("%.2f GB", kvMemory)) + "\n")
	s.WriteString("  Total: " + valueStyle.Render(fmt.Sprintf("%.2f GB", totalMemory)) + "\n")
	s.WriteString("  Per User: " + valueStyle.Render(fmt.Sprintf("%.2f GB", kvMemory/float64(m.users))) + "\n")

	return s.String()
}

func (m Model) renderModelInfo() string {
	var s strings.Builder

	// Model metadata
	s.WriteString("Model ID: " + m.modelInfo.ModelID + "\n")
	s.WriteString("Author: " + m.modelInfo.Author + "\n")
	s.WriteString("Parameters: " + valueStyle.Render(fmt.Sprintf("%.2fB", m.modelInfo.ParametersB)) + "\n")

	// Usage statistics
	s.WriteString("\nUsage Statistics:\n")
	s.WriteString("Downloads: " + valueStyle.Render(fmt.Sprint(m.modelInfo.Downloads)) + "\n")
	s.WriteString("Likes: " + valueStyle.Render(fmt.Sprint(m.modelInfo.Likes)) + "\n")

	// Timing information
	s.WriteString("\nLast Updated: " + m.modelInfo.FetchedAt.Format("2006-01-02 15:04:05") + "\n")

	return s.String()
}

func (m Model) renderConfigurationOptions() string {
	var s strings.Builder

	// User count options
	s.WriteString("\nUsers (+/-):")
	for i, count := range userCounts {
		if i > 0 {
			s.WriteString(" |")
		}
		if count == m.users {
			s.WriteString(" " + selectedStyle.Render(fmt.Sprint(count)))
		} else {
			s.WriteString(" " + fmt.Sprint(count))
		}
	}

	// Context length options
	s.WriteString("\nContext (c):")
	for i, length := range contextLengths {
		if i > 0 {
			s.WriteString(" |")
		}
		if length == m.contextLen {
			s.WriteString(" " + selectedStyle.Render(formatContextLength(length)))
		} else {
			s.WriteString(" " + formatContextLength(length))
		}
	}

	return s.String()
}
