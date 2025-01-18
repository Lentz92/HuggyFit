// internal/tui/tui.go

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lentz92/huggyfit/internal/cache"
	"github.com/Lentz92/huggyfit/internal/calculator"
	"github.com/Lentz92/huggyfit/internal/models"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7B2FBE")).
			Padding(0, 1)

	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			Width(40)

	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Width(50)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Background(lipgloss.Color("#2D2D2D")).
			Padding(0, 1).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B2FF")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	contextLengths = []int{2048, 4096, 8192, 16384, 32768}
	userCounts     = []int{1, 2, 4, 8, 16, 32}
)

type cacheUpdateMsg struct {
	key    cache.CacheKey
	memory float64
}

type Model struct {
	modelIDs              []string
	cursor                int
	modelInfo             *models.ModelInfo
	loading               bool
	err                   error
	spinner               spinner.Model
	textInput             textinput.Model
	searchMode            bool
	quitting              bool
	users                 int
	contextLen            int
	cache                 *cache.Cache
	showHelp              bool
	activeTab             int // 0: memory, 1: details
	cacheOperationPending bool
}

func getNextContextLength(current int) int {
	for i, length := range contextLengths {
		if current <= length {
			if i+1 < len(contextLengths) {
				return contextLengths[i+1]
			}
			return contextLengths[0]
		}
	}
	return contextLengths[0]
}

func getNextUserCount(current int) int {
	for i, count := range userCounts {
		if current <= count {
			if i+1 < len(userCounts) {
				return userCounts[i+1]
			}
			return userCounts[0]
		}
	}
	return userCounts[0]
}

func getPrevUserCount(current int) int {
	for i := len(userCounts) - 1; i >= 0; i-- {
		if current >= userCounts[i] {
			if i > 0 {
				return userCounts[i-1]
			}
			return userCounts[len(userCounts)-1]
		}
	}
	return userCounts[len(userCounts)-1]
}

func InitialModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	ti := textinput.New()
	ti.Placeholder = "Search models..."
	ti.CharLimit = 156
	ti.Width = 40
	ti.Prompt = "🔍 "

	return Model{
		spinner:    s,
		loading:    true,
		textInput:  ti,
		users:      userCounts[0],
		contextLen: contextLengths[1],
		cache:      cache.NewCache(24 * time.Hour),
		activeTab:  0,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchInitialModels,
	)
}

type errMsg error
type modelListMsg []string
type modelInfoMsg *models.ModelInfo

func fetchInitialModels() tea.Msg {
	modelIDs, err := models.FetchModelList()
	if err != nil {
		return errMsg(err)
	}
	return modelListMsg(modelIDs)
}

func performSearch(query string) tea.Cmd {
	return func() tea.Msg {
		modelIDs, err := models.SearchModelList(query)
		if err != nil {
			return errMsg(err)
		}
		return modelListMsg(modelIDs)
	}
}

func fetchModelInfo(modelID string) tea.Cmd {
	return func() tea.Msg {
		info, err := models.FetchModelInfo(modelID)
		if err != nil {
			return errMsg(err)
		}
		return modelInfoMsg(info)
	}
}

func performCacheOperation(m *Model, key cache.CacheKey, parameters float64) tea.Cmd {
	return func() tea.Msg {
		memory := m.cache.GetOrCalculateKVCache(key, parameters, false)
		return cacheUpdateMsg{key: key, memory: memory}
	}
}

func (m Model) calculateKVCache(dtype calculator.DataType) float64 {
	if m.modelInfo == nil {
		return 0
	}

	key := cache.CacheKey{
		ModelID:    m.modelInfo.ModelID,
		Users:      m.users,
		ContextLen: m.contextLen,
		DataType:   dtype,
	}

	// Only perform calculation if it's already in cache
	if value, exists := m.cache.GetKVCache(key); exists {
		return value
	}

	// Return 0 if calculation is pending
	return 0
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		if m.searchMode {
			switch msg.String() {
			case "esc":
				m.searchMode = false
				m.textInput.Blur()
				return m, nil
			case "enter":
				m.loading = true
				m.searchMode = false
				m.textInput.Blur()
				// Cancel any pending cache operations when starting a new search
				m.cacheOperationPending = false
				return m, performSearch(m.textInput.Value())
			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "q":
			m.quitting = true
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
		case "tab":
			if m.modelInfo != nil {
				m.activeTab = (m.activeTab + 1) % 2
			}
		case "/":
			// Allow entering search mode even if cache operations are pending
			m.searchMode = true
			m.textInput.Focus()
			// Reset the cache operation state when entering search
			m.cacheOperationPending = false
			return m, textinput.Blink
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.modelIDs)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.modelIDs) > 0 {
				m.loading = true
				return m, fetchModelInfo(m.modelIDs[m.cursor])
			}
		case "home":
			m.cursor = 0
		case "end":
			m.cursor = len(m.modelIDs) - 1
		case "pgup":
			m.cursor -= 10
			if m.cursor < 0 {
				m.cursor = 0
			}
		case "pgdown":
			m.cursor += 10
			if m.cursor >= len(m.modelIDs) {
				m.cursor = len(m.modelIDs) - 1
			}
		case "+":
			if m.modelInfo != nil {
				m.users = getNextUserCount(m.users)
			}
		case "-":
			if m.modelInfo != nil {
				m.users = getPrevUserCount(m.users)
			}
		case "c":
			if m.modelInfo != nil {
				m.contextLen = getNextContextLength(m.contextLen)
			}
		}

	case modelListMsg:
		m.loading = false
		m.modelIDs = []string(msg)
		m.modelInfo = nil
		m.cursor = 0
		m.err = nil
		m.cacheOperationPending = false // Reset cache operation state
		return m, nil

	case modelInfoMsg:
		m.loading = false
		m.modelInfo = msg
		m.err = nil
		m.cacheOperationPending = true

		// Trigger cache operations for all data types
		var cmds []tea.Cmd
		for _, dtype := range []calculator.DataType{calculator.Float16, calculator.Int8, calculator.Int4} {
			key := cache.CacheKey{
				ModelID:    m.modelInfo.ModelID,
				Users:      m.users,
				ContextLen: m.contextLen,
				DataType:   dtype,
			}
			cmds = append(cmds, performCacheOperation(&m, key, m.modelInfo.ParametersB))
		}
		return m, tea.Batch(cmds...)

	case cacheUpdateMsg:
		if m.cacheOperationPending {
			// Only check remaining operations if we're still pending
			remainingOps := 0
			for _, dtype := range []calculator.DataType{calculator.Float16, calculator.Int8, calculator.Int4} {
				key := cache.CacheKey{
					ModelID:    m.modelInfo.ModelID,
					Users:      m.users,
					ContextLen: m.contextLen,
					DataType:   dtype,
				}
				if _, exists := m.cache.GetKVCache(key); !exists {
					remainingOps++
				}
			}
			if remainingOps == 0 {
				m.cacheOperationPending = false
			}
		}
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg
		m.cacheOperationPending = false // Reset cache operation state on error
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "Thanks for using HuggyFit!\n"
	}

	var s strings.Builder
	s.WriteString(titleStyle.Render("🤗 HuggyFit - GPU Memory Calculator") + "\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("⚠️  Error: %v", m.err)) + "\n")
	}

	if m.searchMode {
		s.WriteString(m.textInput.View() + "\n\n")
	}

	if m.loading {
		return s.String() + fmt.Sprintf("%s Loading...\n", m.spinner.View())
	}

	mainDisplay := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.renderModelList(),
		m.renderModelDetails(),
	)
	s.WriteString(mainDisplay + "\n")

	if m.showHelp {
		s.WriteString(m.renderHelp())
	} else {
		s.WriteString(m.renderControls())
	}

	return s.String()
}

func (m Model) renderModelList() string {
	var s strings.Builder

	if len(m.modelIDs) == 0 {
		return listStyle.Render("No models found")
	}

	const itemsPerPage = 10
	currentPage := m.cursor / itemsPerPage
	start := currentPage * itemsPerPage
	end := start + itemsPerPage
	if end > len(m.modelIDs) {
		end = len(m.modelIDs)
	}

	s.WriteString("Available Models\n")
	s.WriteString(strings.Repeat("─", 38))
	s.WriteString("\n")

	for i := start; i < end; i++ {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		modelID := m.modelIDs[i]
		if len(modelID) > 35 {
			modelID = modelID[:32] + "..."
		}

		paddedModel := fmt.Sprintf("%-35s", modelID)

		if m.cursor == i {
			s.WriteString(fmt.Sprintf("%s %s\n", cursor, selectedStyle.Render(paddedModel)))
		} else {
			s.WriteString(fmt.Sprintf("%s %s\n", cursor, paddedModel))
		}
	}

	s.WriteString(strings.Repeat("─", 38))
	s.WriteString("\n")
	currentPageNum := (start / itemsPerPage) + 1
	totalPages := (len(m.modelIDs) + itemsPerPage - 1) / itemsPerPage
	s.WriteString(fmt.Sprintf("Page %d of %d (%d models)",
		currentPageNum,
		totalPages,
		len(m.modelIDs)))

	return listStyle.Render(s.String())
}

func (m Model) renderModelDetails() string {
	if m.modelInfo == nil {
		return detailStyle.Render("Select a model to view details")
	}

	var s strings.Builder

	// Tabs
	tabs := []string{"Memory Requirements", "Model Details"}
	for i, tab := range tabs {
		if i == m.activeTab {
			s.WriteString(selectedStyle.Render("(" + tab + ")"))
		} else {
			s.WriteString("(" + tab + ")")
		}
		if i < len(tabs)-1 {
			s.WriteString(" ")
		}
	}
	s.WriteString("\n\n")

	if m.activeTab == 0 {
		s.WriteString(m.renderMemoryDetails())
	} else {
		s.WriteString(m.renderModelInfo())
	}

	return detailStyle.Render(s.String())
}

func (m Model) renderMemoryDetails() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("Model: %s\n", m.modelInfo.ModelID))
	s.WriteString(fmt.Sprintf("\nConfiguration:\n"))
	s.WriteString(fmt.Sprintf("- Users: %d\n", m.users))
	s.WriteString(fmt.Sprintf("- Context Length: %dk tokens\n", m.contextLen/1024))

	dtypes := []calculator.DataType{calculator.Float16, calculator.Int8, calculator.Int4}
	for _, dtype := range dtypes {
		baseMemory, _ := calculator.CalculateGPUMemory(m.modelInfo.ParametersB, dtype)
		kvMemory := m.calculateKVCache(dtype)
		totalMemory := baseMemory + kvMemory

		s.WriteString(fmt.Sprintf("\n%s:\n", dtype))
		s.WriteString(fmt.Sprintf("  Base: %.2f GB\n", baseMemory))
		s.WriteString(fmt.Sprintf("  KV Cache: %.2f GB\n", kvMemory))
		s.WriteString(fmt.Sprintf("  Total: %.2f GB\n", totalMemory))
		s.WriteString(fmt.Sprintf("  Per User: %.2f GB\n", kvMemory/float64(m.users)))
	}

	return s.String()
}

func (m Model) renderModelInfo() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("Model ID: %s\n", m.modelInfo.ModelID))
	s.WriteString(fmt.Sprintf("Author: %s\n", m.modelInfo.Author))
	s.WriteString(fmt.Sprintf("Parameters: %.2fB\n", m.modelInfo.ParametersB))
	s.WriteString(fmt.Sprintf("Downloads: %d\n", m.modelInfo.Downloads))
	s.WriteString(fmt.Sprintf("Likes: %d\n", m.modelInfo.Likes))
	s.WriteString(fmt.Sprintf("Last Updated: %s\n", m.modelInfo.FetchedAt.Format("2006-01-02 15:04:05")))

	return s.String()
}

func (m Model) renderControls() string {
	var s strings.Builder

	// Navigation controls
	s.WriteString("\nNavigation:\n")
	s.WriteString(helpStyle.Render("↑/↓ or j/k: Navigate • Enter: Select • /: Search • Tab: Switch view • ?: Help • q: Quit\n"))

	// Show configuration controls only when a model is selected
	if m.modelInfo != nil {
		// User count options
		s.WriteString("\nUsers (+/-):")
		for i, count := range userCounts {
			if i > 0 {
				s.WriteString(" |")
			}
			if count == m.users {
				s.WriteString(" " + selectedStyle.Render(fmt.Sprintf("%d", count)))
			} else {
				s.WriteString(fmt.Sprintf(" %d", count))
			}
		}

		// Context length options
		s.WriteString("\nContext (c):")
		for i, length := range contextLengths {
			if i > 0 {
				s.WriteString(" |")
			}
			if length == m.contextLen {
				s.WriteString(" " + selectedStyle.Render(fmt.Sprintf("%dk", length/1024)))
			} else {
				s.WriteString(fmt.Sprintf(" %dk", length/1024))
			}
		}
	}

	return s.String()
}

func (m Model) renderHelp() string {
	var s strings.Builder

	helpContent := []struct {
		category string
		items    []struct{ key, desc string }
	}{
		{
			category: "Navigation",
			items: []struct{ key, desc string }{
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
			items: []struct{ key, desc string }{
				{"+/-", "Adjust user count"},
				{"c", "Cycle context length"},
			},
		},
		{
			category: "Display",
			items: []struct{ key, desc string }{
				{"?", "Toggle help"},
			},
		},
	}

	s.WriteString("\nHelp:\n")
	for _, category := range helpContent {
		s.WriteString(fmt.Sprintf("\n%s:\n", category.category))
		for _, item := range category.items {
			s.WriteString(fmt.Sprintf("  %-12s: %s\n", selectedStyle.Render(item.key), item.desc))
		}
	}

	s.WriteString(helpStyle.Render("\nPress ? to hide help"))
	return s.String()
}
