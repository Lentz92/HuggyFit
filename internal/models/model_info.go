// internal/models/model_info.go

package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const huggingFaceAPI = "https://huggingface.co/api/models/%s"

// HFResponse represents the HuggingFace API response structure
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

// ModelInfo contains processed model information
type ModelInfo struct {
	ModelID     string
	Author      string
	ParametersB float64
	Downloads   int
	Likes       int
	FetchedAt   time.Time
}

// FetchModelInfo retrieves model information from HuggingFace
func FetchModelInfo(modelID string) (*ModelInfo, error) {
	if modelID == "" {
		return nil, fmt.Errorf("model ID cannot be empty")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request to HuggingFace API
	url := fmt.Sprintf(huggingFaceAPI, modelID)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch model info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var hfResp HFResponse
	if err := json.Unmarshal(body, &hfResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert parameter count to billions
	paramCount := float64(hfResp.Safetensors.Total) / 1e9

	if paramCount == 0 {
		return nil, fmt.Errorf("could not determine parameter count for model: %s", modelID)
	}

	return &ModelInfo{
		ModelID:     hfResp.ModelID,
		Author:      hfResp.Author,
		ParametersB: paramCount,
		Downloads:   hfResp.Downloads,
		Likes:       hfResp.Likes,
		FetchedAt:   time.Now(),
	}, nil
}
