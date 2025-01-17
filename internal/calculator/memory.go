// internal/calculator/memory.go

package calculator

import "fmt"

// DataType represents supported model data types
type DataType string

const (
	// Standard names
	Int4    DataType = "int4"
	Int8    DataType = "int8"
	Float16 DataType = "float16"

	// Common aliases
	Q4  DataType = "q4"  // Alias for int4
	Q8  DataType = "q8"  // Alias for int8
	F16 DataType = "f16" // Alias for float16
)

// BytesPerType maps data types to their byte sizes
var BytesPerType = map[DataType]float64{
	Int4:    0.5,
	Int8:    1.0,
	Float16: 2.0,
	// Aliases map to the same values
	Q4:  0.5,
	Q8:  1.0,
	F16: 2.0,
}

// NormalizeDataType converts common names to standard types
func NormalizeDataType(dtype DataType) DataType {
	switch dtype {
	case Q4:
		return Int4
	case Q8:
		return Int8
	case F16:
		return Float16
	default:
		return dtype
	}
}

// CalculateGPUMemory calculates the GPU memory required for serving a Large Language Model (LLM).
// Formula: M = (P * 4B) / (32 / Q) * 1.18
// where:
// - M is the GPU memory in Gigabytes
// - P is the number of parameters in billions
// - 4B represents 4 bytes per parameter
// - 32 represents bits in 4 bytes
// - Q is the quantization bits (e.g., 16, 8, or 4 bits)
// - 1.18 represents ~18% overhead for additional GPU memory requirements
func CalculateGPUMemory(parameters float64, dtype DataType) (float64, error) {
	const (
		bytesPerParameter = 4    // 4B represents 4 bytes per parameter
		bitsInByte        = 8    // 8 bits in a byte
		bitsInWord        = 32   // 32-bit word size
		overheadFactor    = 1.18 // ~18% overhead for additional GPU memory requirements
	)

	bytes, ok := BytesPerType[dtype]
	if !ok {
		return 0, ErrUnsupportedDataType{dtype}
	}

	// Calculate quantization bits (Q) from bytes
	quantizationBits := bytes * bitsInByte

	// M = (P * 4B) / (32 / Q) * 1.18
	memory := (parameters * float64(bytesPerParameter)) / (float64(bitsInWord) / quantizationBits) * overheadFactor

	return round(memory, 2), nil
}

// ValidateDataType checks if the provided data type is supported
func ValidateDataType(dtype DataType) bool {
	_, exists := BytesPerType[dtype]
	return exists
}

// GetSupportedTypes returns a list of supported data types
func GetSupportedTypes() []DataType {
	types := make([]DataType, 0, len(BytesPerType))
	for dtype := range BytesPerType {
		types = append(types, dtype)
	}
	return types
}

// ErrUnsupportedDataType represents an error for unsupported data types
type ErrUnsupportedDataType struct {
	DataType DataType
}

func (e ErrUnsupportedDataType) Error() string {
	return fmt.Sprintf("unsupported data type: %s", e.DataType)
}

// round rounds a float64 to a specified number of decimal places
func round(num float64, decimals int) float64 {
	multiplier := 1.0
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}
	return float64(int(num*multiplier+0.5)) / multiplier
}
