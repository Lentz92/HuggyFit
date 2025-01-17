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
	verbose := flag.Bool("verbose", false, "Show detailed model information")
	help := flag.Bool("help", false, "Show help message")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "HuggyFit - GPU Memory Calculator for HuggingFace Models\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSupported data types:\n")
		fmt.Fprintf(os.Stderr, "  - float16 (or f16)\n")
		fmt.Fprintf(os.Stderr, "  - int8 (or q8)\n")
		fmt.Fprintf(os.Stderr, "  - int4 (or q4)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Basic usage\n")
		fmt.Fprintf(os.Stderr, "  %s -model Qwen/Qwen2.5-0.5B\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n  # With specific quantization\n")
		fmt.Fprintf(os.Stderr, "  %s -model Qwen/Qwen2.5-0.5B -dtype q4\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n  # Show detailed information\n")
		fmt.Fprintf(os.Stderr, "  %s -model Qwen/Qwen2.5-0.5B -verbose\n", os.Args[0])
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

	// Calculate memory requirements
	memory, err := calculator.CalculateGPUMemory(modelInfo.ParametersB, dtype)
	if err != nil {
		log.Fatalf("Error calculating GPU memory: %v", err)
	}

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
		fmt.Printf("- Required GPU Memory: %.2f GB\n", memory)
	} else {
		fmt.Printf("Estimated GPU memory requirement for %s: %.2f GB (%s)\n",
			modelInfo.ModelID, memory, dtype)
	}
}
