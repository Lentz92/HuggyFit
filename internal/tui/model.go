// internal/tui/model.go

package tui

import (
	"time"

	"github.com/Lentz92/huggyfit/internal/cache"
	"github.com/Lentz92/huggyfit/internal/calculator"
	"github.com/Lentz92/huggyfit/internal/models"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the application state
type Model struct {
	// Core data
	modelIDs  []string
	modelInfo *models.ModelInfo
	cursor    int

	// UI Components
	spinner   spinner.Model
	textInput textinput.Model

	// UI State
	loading               bool
	err                   error
	searchMode            bool
	quitting              bool
	showHelp              bool
	activeTab             int
	cacheOperationPending bool

	// Configuration
	users      int
	contextLen int
	cache      *cache.Cache

	// Terminal size fields
	width  int
	height int
}

// InitialModel creates a new model with default settings
func InitialModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle()

	ti := textinput.New()
	ti.Placeholder = "Search models..."
	ti.CharLimit = 156
	ti.Width = maxSearchWidth
	ti.Prompt = "🔍 "

	return Model{
		// Initialize UI components
		spinner:   s,
		textInput: ti,

		// Set default state
		loading:    true,
		activeTab:  0,
		users:      userCounts[0],
		contextLen: contextLengths[1],
		cache:      cache.NewCache(24 * time.Hour),

		// Initialize with default dimensions
		width:  getMainContentWidth(),
		height: getMainContentHeight(),
	}
}

// Init returns the initial command for the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchInitialModels,
	)
}

// calculateKVCache calculates KV cache memory requirements
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

	// Return cached value if available
	if value, exists := m.cache.GetKVCache(key); exists {
		return value
	}

	// Return 0 if calculation is pending
	return 0
}

// isModelSelected returns whether a model is currently selected
func (m Model) isModelSelected() bool {
	return m.modelInfo != nil
}

// hasModels returns whether there are any models in the list
func (m Model) hasModels() bool {
	return len(m.modelIDs) > 0
}

// Command generators
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
