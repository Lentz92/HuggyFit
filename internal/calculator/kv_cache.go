// internal/calculator/kv_cache.go

package calculator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ModelConfig represents the relevant fields from config.json
type ModelConfig struct {
	HiddenSize        int `json:"hidden_size"`
	NumAttentionHeads int `json:"num_attention_heads"`
	NumHiddenLayers   int `json:"num_hidden_layers"`
	NumKeyValueHeads  int `json:"num_key_value_heads"`
}

// KVCacheParams holds parameters for KV cache calculation
type KVCacheParams struct {
	Users         int
	ContextLength int
	DataType      DataType
	Config        *ModelConfig
}

// FetchModelConfig retrieves the model's configuration from HuggingFace
func FetchModelConfig(modelID string) (*ModelConfig, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("https://huggingface.co/%s/raw/main/config.json", modelID)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch model config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		return nil, fmt.Errorf("\n%s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read config response: %w", err)
	}

	var config ModelConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Handle models that don't specify num_key_value_heads
	if config.NumKeyValueHeads == 0 {
		config.NumKeyValueHeads = config.NumAttentionHeads
	}

	return &config, nil
}

// CalculateKVCache computes memory required for KV cache per user
func CalculateKVCache(params KVCacheParams) (float64, error) {
	if params.Config == nil {
		return 0, fmt.Errorf("model config is required for KV cache calculation")
	}

	bytes, ok := BytesPerType[params.DataType]
	if !ok {
		return 0, ErrUnsupportedDataType{params.DataType}
	}

	// KV Cache formula:
	// Memory = 2 * num_layers * seq_len * (hidden_size/num_attn_heads * num_kv_heads) * 2 * bytes_per_param * num_users
	kvSize := float64(2 * params.Config.NumHiddenLayers * params.ContextLength *
		(params.Config.HiddenSize / params.Config.NumAttentionHeads * params.Config.NumKeyValueHeads) * 2)

	// Convert to GB
	memoryGB := (kvSize * bytes) / (1024 * 1024 * 1024)

	// Apply per-user scaling
	totalMemoryGB := memoryGB * float64(params.Users)

	return round(totalMemoryGB, 2), nil
}

// EstimateKVCache provides an estimation for gated models
func EstimateKVCache(parameterCount float64, users, contextLength int, dtype DataType) float64 {
	// Estimation based on model size:
	// Small (< 7B): ~0.5GB per 1k tokens
	// Medium (7-20B): ~1GB per 1k tokens
	// Large (> 20B): ~2GB per 1k tokens
	var memoryPerUser float64
	switch {
	case parameterCount < 7:
		memoryPerUser = 0.5
	case parameterCount < 20:
		memoryPerUser = 1.0
	default:
		memoryPerUser = 2.0
	}

	// Scale by context length (normalized to 1k tokens)
	memoryPerUser *= float64(contextLength) / 1000.0

	// Apply dtype scaling
	bytes, _ := BytesPerType[dtype]
	dtypeScale := bytes / BytesPerType[Float16] // normalize to FP16
	memoryPerUser *= dtypeScale

	return round(memoryPerUser*float64(users), 2)
}
