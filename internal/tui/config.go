// internal/tui/config.go

package tui

import (
	"fmt"

	"github.com/Lentz92/huggyfit/internal/calculator"
)

const (
	// UI Constants
	itemsPerPage   = 10
	maxSearchWidth = 40
)

// Default window dimensions
const (
	defaultWidth  = 100
	defaultHeight = 30
)

// Predefined context lengths in tokens
var contextLengths = []int{
	2048,  // 2k
	4096,  // 4k
	8192,  // 8k
	16384, // 16k
	32768, // 32k
}

// Predefined user count options
var userCounts = []int{1, 2, 4, 8, 16, 32}

// Supported data types for memory calculation
var dataTypes = []calculator.DataType{
	calculator.Float16,
	calculator.Int8,
	calculator.Int4,
}

// getMainContentWidth returns the desired width for the main content area
func getMainContentWidth() int {
	return defaultWidth
}

// getMainContentHeight returns the desired height for the main content area
func getMainContentHeight() int {
	return defaultHeight
}

// getNextContextLength returns the next available context length
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

// getNextUserCount returns the next available user count
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

// getPrevUserCount returns the previous available user count
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

// formatContextLength formats a context length for display
func formatContextLength(length int) string {
	return fmt.Sprintf("%dk", length/1024)
}

// getCurrentPage calculates the current page number based on cursor position
func getCurrentPage(cursor int) int {
	return cursor / itemsPerPage
}

// getPageBounds returns the start and end indices for the current page
func getPageBounds(cursor, totalItems int) (start, end int) {
	currentPage := getCurrentPage(cursor)
	start = currentPage * itemsPerPage
	end = start + itemsPerPage
	if end > totalItems {
		end = totalItems
	}
	return start, end
}
