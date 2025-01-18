// internal/tui/tui.go

package tui

import (
	"fmt"
	"strings"

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
)

type Model struct {
	modelIDs   []string
	cursor     int
	modelInfo  *models.ModelInfo
	loading    bool
	err        error
	spinner    spinner.Model
	textInput  textinput.Model
	searchMode bool
	quitting   bool
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
		spinner:   s,
		loading:   true,
		textInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchInitialModels,
	)
}

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

type errMsg error
type modelListMsg []string
type modelInfoMsg *models.ModelInfo

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle ctrl+c globally for emergency exit
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		// Handle other keys based on mode
		if m.searchMode {
			switch msg.String() {
			case "esc":
				m.searchMode = false
				m.textInput.Blur()
				return m, nil
			case "enter":
				m.loading = true
				m.searchMode = false // Exit search mode automatically
				m.textInput.Blur()   // Blur the input
				return m, performSearch(m.textInput.Value())
			default:
				// Let all other keys go to the text input when in search mode
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		} else {
			// Normal mode key handling
			switch msg.String() {
			case "q":
				m.quitting = true
				return m, tea.Quit
			case "/":
				m.searchMode = true
				m.textInput.Focus()
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
			}
		}

	case modelListMsg:
		m.loading = false
		m.modelIDs = []string(msg)
		m.modelInfo = nil
		m.cursor = 0
		m.err = nil // Clear any previous errors
		return m, nil

	case modelInfoMsg:
		m.loading = false
		m.modelInfo = msg
		m.err = nil // Clear any previous errors
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg
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

	// Show error as a banner if present
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
	s.WriteString("\nNavigate: ↑/↓ or j/k • Select: Enter • Search: / • Quit: q")

	return s.String()
}

func (m Model) renderModelList() string {
	var s strings.Builder

	if len(m.modelIDs) == 0 {
		return listStyle.Render("No models found")
	}

	// Calculate pagination
	const itemsPerPage = 10
	currentPage := m.cursor / itemsPerPage
	start := currentPage * itemsPerPage
	end := start + itemsPerPage
	if end > len(m.modelIDs) {
		end = len(m.modelIDs)
	}

	// Add header
	s.WriteString("Available Models\n")
	s.WriteString(strings.Repeat("─", 38)) // Separator line
	s.WriteString("\n")

	// Format each model entry for current page
	for i := start; i < end; i++ {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		// Format model ID to fit within the list
		modelID := m.modelIDs[i]
		if len(modelID) > 35 {
			modelID = modelID[:32] + "..."
		}

		// Add padding to align items
		paddedModel := fmt.Sprintf("%-35s", modelID)

		// Style the selected item
		if m.cursor == i {
			s.WriteString(fmt.Sprintf("%s %s\n", cursor,
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("#74B2FF")).
					Bold(true).
					Render(paddedModel)))
		} else {
			s.WriteString(fmt.Sprintf("%s %s\n", cursor, paddedModel))
		}
	}

	// Add footer with pagination info
	s.WriteString(strings.Repeat("─", 38)) // Separator line
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
	s.WriteString(fmt.Sprintf("Model: %s\n", m.modelInfo.ModelID))
	s.WriteString(fmt.Sprintf("Author: %s\n", m.modelInfo.Author))
	s.WriteString(fmt.Sprintf("Parameters: %.2fB\n", m.modelInfo.ParametersB))
	s.WriteString(fmt.Sprintf("Downloads: %d\n", m.modelInfo.Downloads))
	s.WriteString(fmt.Sprintf("Likes: %d\n", m.modelInfo.Likes))

	s.WriteString("\nEstimated GPU Memory Requirements:\n")
	params := m.modelInfo.ParametersB
	s.WriteString(fmt.Sprintf("float16: %.2f GB\n", params*2.0*1.18))
	s.WriteString(fmt.Sprintf("int8: %.2f GB\n", params*1.0*1.18))
	s.WriteString(fmt.Sprintf("int4: %.2f GB\n", params*0.5*1.18))

	return detailStyle.Render(s.String())
}
