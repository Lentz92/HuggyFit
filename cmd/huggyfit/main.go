// cmd/huggyfit/main.go

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Lentz92/huggyfit/internal/calculator"
	"github.com/Lentz92/huggyfit/internal/models"
)

func main() {
	// Setup command line flags
	modelID := flag.String("model", "", "HuggingFace model ID (e.g., Qwen/Qwen2.5-0.5B)")
	dtypeStr := flag.String("dtype", string(calculator.Float16),
		"Data type for model loading (float16/f16, int8/q8, int4/q4)")
	users := flag.Int("users", 1, "Number of concurrent users")
	contextLen := flag.Int("context", 4096, "Context length per user")
	estimateKV := flag.Bool("estimate-kv", false, "Use estimation for KV cache calculation")
	verbose := flag.Bool("verbose", false, "Show detailed model information")
	help := flag.Bool("help", false, "Show help message")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "HuggyFit - GPU Memory Calculator for HuggingFace Models\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Basic usage with concurrent users\n")
		fmt.Fprintf(os.Stderr, "  %s -model Qwen/Qwen2.5-0.5B -users 4\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n  # With specific context length\n")
		fmt.Fprintf(os.Stderr, "  %s -model Qwen/Qwen2.5-0.5B -users 2 -context 8192\n", os.Args[0])
	}
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *modelID == "" {
		fmt.Println("Error: model ID is required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate and normalize data type
	dtype := calculator.NormalizeDataType(calculator.DataType(strings.ToLower(*dtypeStr)))
	if !calculator.ValidateDataType(dtype) {
		log.Printf("Error: unsupported data type: %s\n", dtype)
		log.Printf("Supported types: float16/f16, int8/q8, int4/q4\n")
		os.Exit(1)
	}

	// Fetch model information
	modelInfo, err := models.FetchModelInfo(*modelID)
	if err != nil {
		log.Fatalf("Error fetching model information: %v", err)
	}

	// Calculate base memory requirements
	baseMemory, err := calculator.CalculateGPUMemory(modelInfo.ParametersB, dtype)
	if err != nil {
		log.Fatalf("Error calculating base GPU memory: %v", err)
	}

	var kvMemory float64
	if !*estimateKV {
		// Try to fetch model config for precise KV cache calculation
		config, err := calculator.FetchModelConfig(*modelID)
		if err == nil {
			kvParams := calculator.KVCacheParams{
				Users:         *users,
				ContextLength: *contextLen,
				DataType:      dtype,
				Config:        config,
			}
			kvMemory, err = calculator.CalculateKVCache(kvParams)
			if err != nil {
				log.Printf("Warning: Failed to calculate precise KV cache: %v\n", err)
				log.Printf("Falling back to estimation...\n")
				*estimateKV = true
			}
		} else {
			log.Printf("Warning: Failed to fetch model config: %v\n", err)
			log.Printf("Falling back to estimation...\n")
			*estimateKV = true
		}
	}

	if *estimateKV {
		kvMemory = calculator.EstimateKVCache(modelInfo.ParametersB, *users, *contextLen, dtype)
	}

	totalMemory := baseMemory + kvMemory

	// Display results
	if *verbose {
		fmt.Printf("\nModel Information:\n")
		fmt.Printf("- Model ID: %s\n", modelInfo.ModelID)
		fmt.Printf("- Author: %s\n", modelInfo.Author)
		fmt.Printf("- Parameters: %.2fB\n", modelInfo.ParametersB)
		fmt.Printf("- Downloads: %d\n", modelInfo.Downloads)
		fmt.Printf("- Likes: %d\n", modelInfo.Likes)
		fmt.Printf("\nMemory Requirements:\n")
		fmt.Printf("- Data Type: %s\n", dtype)
		fmt.Printf("- Base Model Memory: %.2f GB\n", baseMemory)
		fmt.Printf("- KV Cache Memory: %.2f GB (%s)\n",
			kvMemory,
			map[bool]string{true: "estimated", false: "precise"}[*estimateKV])
		fmt.Printf("- KV Cache Per User: %.2f GB\n", kvMemory/float64(*users))
		fmt.Printf("- Total GPU Memory: %.2f GB\n", totalMemory)
		fmt.Printf("- Users: %d\n", *users)
		fmt.Printf("- Context Length: %d tokens\n", *contextLen)
	} else {
		fmt.Printf("Estimated GPU memory requirement for %s:\n", modelInfo.ModelID)
		fmt.Printf("- Total: %.2f GB (%s)\n", totalMemory, dtype)
		fmt.Printf("- Per User: %.2f GB\n", kvMemory/float64(*users))
	}
}
