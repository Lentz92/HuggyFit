// internal/tui/messages.go

package tui

import (
	"github.com/Lentz92/huggyfit/internal/cache"
	"github.com/Lentz92/huggyfit/internal/models"
)

// Message types for the TUI
type errMsg error
type modelListMsg []string
type modelInfoMsg *models.ModelInfo
type cacheUpdateMsg struct {
	key    cache.CacheKey
	memory float64
}
