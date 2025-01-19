# HuggyFit Project Documentation

## Project Overview

HuggyFit is a terminal user interface (TUI) application that helps developers and system administrators calculate GPU memory requirements for Large Language Models (LLMs). The application lets users explore different models, calculate memory requirements for various configurations, and plan deployments.

##### Project Structure

The project is organized in the following structure:

```
HuggyFit/
├── cmd/                    # Application entry points
│   └── huggyfitui/        # Main TUI application
├── internal/              # Private application code
│   ├── cache/            # Caching system
│   ├── calculator/       # Memory calculation logic
│   ├── models/          # Model data structures and API
│   └── tui/             # Terminal UI implementation
└── docs/                # Documentation
```

**Features:**
- Model search and selection from HuggingFace's repository
- GPU memory calculations for different quantization types (4-bit, 8-bit, 16-bit)
- Multi-user scenario planning
- Adjustable context lengths
- Caching system for repeated calculations

**High-Level Architecture:**

[Mermaid diagram section remains exactly as before...]

The architecture consists of:
- **Frontend:** TUI built with Bubble Tea framework using Model-View-Update pattern
- **Core Engine:** Memory calculations for:
  - Model parameters
  - Quantization
  - KV cache
  - Multi-user scaling
- **Caching Layer:** Storage with 24-hour expiration
- **Model Management:** Model data and search
- **External Integration:** HuggingFace API connection

**Goals:**
1. Calculate memory requirements for LLM deployments
2. Show memory impact of different model configurations
3. Calculate memory scaling for multi-user deployments
4. Provide keyboard-driven interface for model exploration

## Core Components

### Command Line Interface (cmd/)
##### cmd/huggyfitui/main.go
The main entry point initializes the HuggyFit TUI application.

**Dependencies:**
- Uses Bubble Tea framework for terminal UI management
- Integrates with internal TUI package for UI logic

**Program Initialization:**
```go
p := tea.NewProgram(
    tui.InitialModel(),
    tea.WithAltScreen(),       // Use alternate screen buffer
    tea.WithMouseCellMotion(), // Enable mouse support
)
```

**Features:**
1. **Alternate Screen Buffer:** Uses `tea.WithAltScreen()` to run in a separate terminal buffer
2. **Mouse Support:** Enables mouse interaction with `tea.WithMouseCellMotion()`
3. **Error Handling:** Provides error messages and exit codes

### Internal Components (internal/)

#### Cache System
##### internal/cache/cache.go
The cache.go file implements a thread-safe caching system for model configurations and memory calculations.

**Core Types:**
```go
type CacheKey struct {
    ModelID    string
    Users      int
    ContextLen int
    DataType   calculator.DataType
}

type Cache struct {
    configs      map[string]*calculator.ModelConfig
    calculations map[CacheKey]float64
    mu           sync.RWMutex
    expiration   time.Duration
}
```

**Features:**
1. **Thread Safety:**
   - Uses sync.RWMutex for concurrent access
   - Uses read/write locks
   - Works with multiple goroutines

2. **Caching:**
   - Model configurations cache (`configs`)
   - KV cache calculations (`calculations`)

3. **Cache Operations:**
   - GetOrCalculateKVCache for retrieval/computation
   - Falls back to estimation if calculation fails
   - Automatically populates on cache misses

#### Calculator
##### internal/calculator/kv_cache.go
The kv_cache.go file implements the key-value cache calculation system for language models.

**Core Types:**
```go
type ModelConfig struct {
    HiddenSize        int `json:"hidden_size"`
    NumAttentionHeads int `json:"num_attention_heads"`
    NumHiddenLayers   int `json:"num_hidden_layers"`
    NumKeyValueHeads  int `json:"num_key_value_heads"`
}

type KVCacheParams struct {
    Users         int
    ContextLength int
    DataType      DataType
    Config        *ModelConfig
}
```

**Features:**
1. **Model Configuration:**
   - Gets model details from HuggingFace
   - Handles different attention head setups
   - Falls back for non-standard models

2. **Memory Calculation:**
   - Calculates KV cache size from model architecture
   - Supports FP16, INT8, INT4 data types
   - Calculates per-user memory

3. **Estimation:**
   - Estimates for gated or unknown models
   - Uses model size for estimates
   - Adjusts for context length

##### internal/calculator/memory.go
The memory.go file implements GPU memory calculations for language models.

**Core Types and Constants:**
```go
type DataType string

const (
    Int4    DataType = "int4"    // 4-bit integer
    Int8    DataType = "int8"    // 8-bit integer
    Float16 DataType = "float16" // 16-bit floating point
    
    // Common aliases for compatibility
    Q4  DataType = "q4"  // Alias for int4
    Q8  DataType = "q8"  // Alias for int8
    F16 DataType = "f16" // Alias for float16
)

var BytesPerType = map[DataType]float64{
    Int4:    0.5,  // 4 bits = 0.5 bytes
    Int8:    1.0,  // 8 bits = 1 byte
    Float16: 2.0,  // 16 bits = 2 bytes
}
```

**Features:**
1. **Memory Calculation:**
   - Uses industry-standard formulas
   - Supports different quantization levels
   - Includes overhead factor (~18%)

2. **Data Type Management:**
   - Type system with aliases
   - Type normalization
   - Type validation

3. **Precision Control:**
   - Rounding behavior
   - Decimal place standardization
   - Error margin management

#### Models
##### internal/models/model_info.go
The model_info.go file handles model information retrieval and processing using the HuggingFace API.

**Core Types:**
```go
type ModelInfo struct {
    ModelID     string
    Author      string
    ParametersB float64
    Downloads   int
    Likes       int
    FetchedAt   time.Time
}

type HFResponse struct {
    ModelID     string `json:"id"`
    Author      string `json:"author"`
    Downloads   int    `json:"downloads"`
    Likes       int    `json:"likes"`
    Safetensors struct {
        Parameters struct {
            BF16 int64 `json:"BF16"`
        } `json:"parameters"`
        Total int64 `json:"total"`
    } `json:"safetensors"`
}
```

**Features:**
1. **API Integration:**
   - HuggingFace API connection
   - Timeout settings
   - Error handling

2. **Data Processing:**
   - Parameter count conversion
   - Response parsing
   - Timestamp tracking

##### internal/models/model_list.go
The model_list.go file manages model list retrieval and search functionality.

**Core Types:**
```go
type ModelListResponse struct {
    ModelID string `json:"id"`
}

type SearchResult struct {
    ModelID string
    Score   int
}
```

**Features:**
1. **Model List Retrieval:**
   - HuggingFace API integration
   - Response parsing
   - Timeout handling

2. **Search System:**
   - Exact substring matching
   - Fuzzy matching
   - Result ranking
   - Case-insensitive matching

#### Terminal UI
##### internal/tui/config.go
The config.go file manages application configuration and UI calculations.

**Configuration Constants:**
```go
const (
    // UI Constants
    itemsPerPage   = 10
    maxSearchWidth = 40

    // Default window dimensions
    defaultWidth  = 100
    defaultHeight = 30
)
```

**Predefined Values:**
1. **Context Lengths:**
   ```go
   var contextLengths = []int{
       2048,  // 2k
       4096,  // 4k
       8192,  // 8k
       16384, // 16k
       32768, // 32k
   }
   ```

2. **User Counts:**
   ```go
   var userCounts = []int{1, 2, 4, 8, 16, 32}
   ```

3. **Data Types:**
   ```go
   var dataTypes = []calculator.DataType{
       calculator.Float16,
       calculator.Int8,
       calculator.Int4,
   }
   ```

##### internal/tui/details_view.go
The details_view.go file implements a tabbed interface showing memory calculations and model information. It includes:

**View Structure:**
```go
func (m Model) renderModelDetails() string {
    if !m.isModelSelected() {
        return detailStyle.Render("Select a model to view details")
    }

    var s strings.Builder
    s.WriteString(m.renderTabs())
    s.WriteString("\n\n")

    // Content based on active tab
    if m.activeTab == 0 {
        s.WriteString(m.renderMemoryDetails())
    } else {
        s.WriteString(m.renderModelInfo())
    }

    return detailStyle.Render(s.String())
}
```

- Memory requirements tab showing base memory, KV cache, and per-user calculations
- Model details tab displaying model metadata and usage statistics
- Tab switching with keyboard controls

##### internal/tui/help_view.go
The help_view.go file implements the help system with keyboard shortcuts and commands. It includes:

**Help Structure:**
```go
type helpSection struct {
    category string
    items    []helpItem
}

type helpItem struct {
    key  string
    desc string
}
```

- Categorized command groups (Navigation, Configuration)
- Keyboard shortcuts with descriptions
- Toggle-based display system
- Overlay presentation

##### internal/tui/list_view.go
The list_view.go file implements the model list display with these features:

**List Management:**
```go
func (m Model) renderModelList() string {
    // Get current page bounds
    start, end := getPageBounds(m.cursor, len(m.modelIDs))

    // Calculate content width
    contentWidth := getListWidth(m.width) - 7 // Account for borders, padding, cursor
}
```

- Pagination with configurable items per page
- Model ID truncation for long names
- Cursor-based selection
- Page information display

##### internal/tui/messages.go
The messages.go file defines the message types for state management:

**Message Types:**
```go
type errMsg error                // Error handling messages
type modelListMsg []string       // Model list updates
type modelInfoMsg *models.ModelInfo // Model details updates
type cacheUpdateMsg struct {      // Cache operation results
    key    cache.CacheKey
    memory float64
}
```

- Error handling messages
- Model data updates
- Cache operation results
- State synchronization messages

##### internal/tui/model.go
The model.go file implements the application state container:

**State Structure:**
```go
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
```

- Core data management
- UI component state
- Configuration settings
- Terminal dimensions tracking

##### internal/tui/styles.go
The styles.go file implements UI styling with Lipgloss:

**Color System:**
```go
var (
    primaryColor   = lipgloss.Color("#7B2FBE")
    secondaryColor = lipgloss.Color("#874BFD")
    textColor      = lipgloss.Color("#FAFAFA")
    errorColor     = lipgloss.Color("#FF0000")
    highlightColor = lipgloss.Color("#74B2FF")
    mutedColor     = lipgloss.Color("#626262")
)
```

- Color definitions
- Component styles
- Layout styles
- Responsive width calculations

##### internal/tui/update.go
The update.go file handles state updates and user input:

**Update System:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyPress(msg)
    case tea.WindowSizeMsg:
        return m.handleWindowSize(msg)
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
```

- Keyboard input handling
- Window resize management
- Model updates
- Error handling
- Loading state management

##### internal/tui/view.go
The view.go file manages the UI layout and rendering:

**View Components:**
- Header with application title
- Main content area with model list and details
- Loading spinner overlay
- Help system overlay
- Error message display

## Technical Decisions

### Design Patterns

The project uses these architectural patterns:

1. **Model-View-Update (MVU/Elm Architecture):**
   - Core TUI architecture in `internal/tui/`
   - State container (`model.go`)
   - Rendering logic (`view.go`)
   - State transitions (`update.go`)

2. **Command Pattern:**
   - Message handling system in `messages.go`
   - Handles async operations
   - Manages state updates
   - Processes UI events

3. **Observer Pattern:**
   - Event handling in TUI components
   - Window resize events
   - User input handling
   - State change notifications

4. **Factory Pattern:**
   - Model initialization in `model.go`
   - Creates model instances
   - Initializes UI components

5. **Strategy Pattern:**
   - Memory calculation system
   - Different quantization strategies
   - Cache calculation methods

6. **Singleton Pattern:**
   - Cache system
   - Shared cache instance
   - Configuration management

7. **Composite Pattern:**
   - UI component hierarchy
   - View composition
   - Layout management

### Performance Considerations

The project includes these optimizations:

1. **Caching System:**
   - 24-hour expiration
   - Thread-safe implementation
   - Two-tier caching:
     * Model configurations
     * KV cache calculations

2. **Memory Management:**
   - String handling with `strings.Builder`
   - Pre-allocated slices
   - Memory allocation optimization
   - Text truncation
   - Garbage collection handling

3. **UI Performance:**
   - Lazy rendering
   - Pagination
   - Width calculations caching
   - Static element reuse
   - Batched state updates

4. **API Integration:**
   - 10-second timeout
   - Connection pooling
   - Response parsing
   - Resource cleanup
   - Error handling

5. **Calculation Optimizations:**
   - Result caching
   - Estimation fallbacks
   - Type system optimization
   - Memory formulas
   - Batch processing

6. **Search System:**
   - Two-phase search:
     * Exact matching
     * Fuzzy matching
   - String operation optimization
   - Result ranking
   - Case handling

7. **Resource Management:**
   - Resource cleanup
   - Terminal buffer management
   - Window resize handling
   - UI update optimization
   - Goroutine management

### Error Handling

The error handling system includes:

1. **Error Types:**
   - Custom error types
   - Domain-specific errors
   - Error wrapping

2. **Error Propagation:**
   - Message system integration
   - Error context
   - Recovery hints

3. **Recovery Strategies:**
   - Estimation fallbacks
   - Default values
   - State recovery
   - Resource cleanup

4. **User Feedback:**
   - Error messages
   - Status indicators
   - Recovery instructions

5. **Validation:**
   - Input checking
   - Data type validation
   - Boundary checking
   - State validation

6. **Network Handling:**
   - Timeout management
   - Retry logic
   - Circuit breaking
   - Rate limiting

## Dependencies

The project uses these libraries:

1. **Bubble Tea Framework Suite:**
   - `github.com/charmbracelet/bubbletea` (v1.2.4)
   - `github.com/charmbracelet/bubbles` (v0.20.0)
   - `github.com/charmbracelet/lipgloss` (v1.0.0)

2. **Search and Matching:**
   - `github.com/sahilm/fuzzy` (v0.1.1)

3. **Supporting Libraries:**
   - `github.com/atotto/clipboard`
   - `github.com/lucasb-eyer/go-colorful`
   - `github.com/mattn/go-runewidth`
   - `github.com/muesli/termenv`

## Future Improvements

Planned UI improvements:

1. **UI Updates:**
   - Fix keyboard shortcuts overlay
   - Improve responsiveness
   - Update layout handling
   - Optimize rendering

2. **Bug Fixes:**
    - Better error handling
