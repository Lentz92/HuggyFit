// internal/models/model_list.go

package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/sahilm/fuzzy"
)

const (
	modelListAPIURL = "https://huggingface.co/api/models"
	defaultLimit    = 20
)

// ModelListResponse represents a simplified model from the models list API
type ModelListResponse struct {
	ModelID string `json:"id"`
}

// SearchResult represents a model with its search relevance score
type SearchResult struct {
	ModelID string
	Score   int
}

// FetchModelList retrieves a list of model IDs from HuggingFace API
func FetchModelList() ([]string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(modelListAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var models []ModelListResponse
	if err := json.Unmarshal(body, &models); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract just the model IDs
	modelIDs := make([]string, len(models))
	for i, model := range models {
		modelIDs[i] = model.ModelID
	}

	return modelIDs, nil
}

// rankModelResults sorts model IDs by relevance to the search query
func rankModelResults(models []string, query string) []string {
	if query == "" {
		return models
	}

	// First, find exact matches
	exactMatches := make([]string, 0)
	fuzzyMatches := make([]SearchResult, 0)

	query = strings.ToLower(query)

	for _, modelID := range models {
		modelLower := strings.ToLower(modelID)

		// Check for exact matches first
		if strings.Contains(modelLower, query) {
			exactMatches = append(exactMatches, modelID)
			continue
		}

		// Use fuzzy matching for non-exact matches
		matches := fuzzy.Find(query, []string{modelID})
		if len(matches) > 0 {
			fuzzyMatches = append(fuzzyMatches, SearchResult{
				ModelID: modelID,
				Score:   matches[0].Score,
			})
		}
	}

	// Sort exact matches by putting most relevant first
	sort.Slice(exactMatches, func(i, j int) bool {
		iLower := strings.ToLower(exactMatches[i])
		jLower := strings.ToLower(exactMatches[j])

		// Prioritize matches at the start of the string
		iIndex := strings.Index(iLower, query)
		jIndex := strings.Index(jLower, query)
		if iIndex != jIndex {
			return iIndex < jIndex
		}

		// If same position, shorter strings first
		return len(exactMatches[i]) < len(exactMatches[j])
	})

	// Sort fuzzy matches by score
	sort.Slice(fuzzyMatches, func(i, j int) bool {
		return fuzzyMatches[i].Score > fuzzyMatches[j].Score
	})

	// Combine results: exact matches first, then fuzzy matches
	result := make([]string, 0, len(models))
	result = append(result, exactMatches...)

	for _, match := range fuzzyMatches {
		result = append(result, match.ModelID)
	}

	return result
}

// SearchModelList searches for model IDs matching the query
func SearchModelList(query string) ([]string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s?search=%s", modelListAPIURL, query)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to search models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var models []ModelListResponse
	if err := json.Unmarshal(body, &models); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract model IDs
	modelIDs := make([]string, len(models))
	for i, model := range models {
		modelIDs[i] = model.ModelID
	}

	// Rank the results
	return rankModelResults(modelIDs, query), nil
}
