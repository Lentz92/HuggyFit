// internal/tui/update.go

package tui

import (
	"github.com/Lentz92/huggyfit/internal/cache"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all state updates based on messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update text input width based on terminal size
		m.textInput.Width = m.width / 2
		return m, nil
	case modelListMsg:
		return m.handleModelList(msg)
	case modelInfoMsg:
		return m.handleModelInfo(msg)
	case cacheUpdateMsg:
		return m.handleCacheUpdate(msg)
	case errMsg:
		return m.handleError(msg)
	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle search mode keys
	if m.searchMode {
		return m.handleSearchModeKeys(msg)
	}

	// Handle normal mode keys
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	case "/":
		return m.enterSearchMode()
	case "?":
		m.toggleHelp()
		return m, nil
	case "tab":
		if m.isModelSelected() {
			m.activeTab = (m.activeTab + 1) % 2
		}
		return m, nil
	}

	// Handle navigation keys
	return m.handleNavigationKeys(msg)
}

// handleSearchModeKeys handles keys while in search mode
func (m Model) handleSearchModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.textInput.Blur()
		return m, nil
	case "enter":
		m.loading = true
		m.searchMode = false
		m.textInput.Blur()
		m.cacheOperationPending = false
		return m, performSearch(m.textInput.Value())
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

// handleNavigationKeys handles model list navigation
func (m Model) handleNavigationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.modelIDs)-1 {
			m.cursor++
		}
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = len(m.modelIDs) - 1
	case "pgup":
		m.cursor = max(0, m.cursor-itemsPerPage)
	case "pgdown":
		m.cursor = min(len(m.modelIDs)-1, m.cursor+itemsPerPage)
	case "enter":
		if m.hasModels() {
			m.loading = true
			return m, fetchModelInfo(m.modelIDs[m.cursor])
		}
	case "+":
		if m.isModelSelected() {
			m.users = getNextUserCount(m.users)
			return m, m.triggerCacheUpdate()
		}
	case "-":
		if m.isModelSelected() {
			m.users = getPrevUserCount(m.users)
			return m, m.triggerCacheUpdate()
		}
	case "c":
		if m.isModelSelected() {
			m.contextLen = getNextContextLength(m.contextLen)
			return m, m.triggerCacheUpdate()
		}
	}
	return m, nil
}

// handleModelList processes model list updates
func (m Model) handleModelList(msg modelListMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	m.modelIDs = []string(msg)
	m.modelInfo = nil
	m.cursor = 0
	m.err = nil
	m.cacheOperationPending = false
	return m, nil
}

// handleModelInfo processes model info updates
func (m Model) handleModelInfo(msg modelInfoMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	m.modelInfo = msg
	m.err = nil
	m.cacheOperationPending = true

	var cmds []tea.Cmd
	for _, dtype := range dataTypes {
		key := cache.CacheKey{
			ModelID:    m.modelInfo.ModelID,
			Users:      m.users,
			ContextLen: m.contextLen,
			DataType:   dtype,
		}
		cmds = append(cmds, performCacheOperation(&m, key, m.modelInfo.ParametersB))
	}
	return m, tea.Batch(cmds...)
}

// handleCacheUpdate processes cache updates
func (m Model) handleCacheUpdate(msg cacheUpdateMsg) (tea.Model, tea.Cmd) {
	// Update cache with the new value
	m.cache.SetKVCache(msg.key, msg.memory)

	// Check if there are any remaining cache operations
	if m.cacheOperationPending {
		remainingOps := 0
		for _, dtype := range dataTypes {
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
		m.cacheOperationPending = remainingOps > 0
	}
	return m, nil
}

// handleError processes error messages
func (m Model) handleError(msg errMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	m.err = msg
	m.cacheOperationPending = false
	return m, nil
}

// handleSpinnerTick updates the spinner animation
func (m Model) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// enterSearchMode prepares the model for search input
func (m Model) enterSearchMode() (tea.Model, tea.Cmd) {
	m.searchMode = true
	m.textInput.Reset()
	m.textInput.Focus()
	m.textInput.Width = m.width / 2 // Ensure width is set
	m.cacheOperationPending = false
	return m, textinput.Blink
}

// triggerCacheUpdate triggers a recalculation of cache values
func (m Model) triggerCacheUpdate() tea.Cmd {
	if m.modelInfo == nil {
		return nil
	}

	var cmds []tea.Cmd
	for _, dtype := range dataTypes {
		key := cache.CacheKey{
			ModelID:    m.modelInfo.ModelID,
			Users:      m.users,
			ContextLen: m.contextLen,
			DataType:   dtype,
		}
		cmds = append(cmds, performCacheOperation(&m, key, m.modelInfo.ParametersB))
	}
	return tea.Batch(cmds...)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
