// Package helpers provides generic utilities for JSON operations and common helper functions.
//
// This package offers type-safe JSON marshaling/unmarshaling operations using Go generics,
// along with validation, formatting, and utility functions for working with JSON data.
//
// Example usage:
//
//	type User struct {
//		ID   int    `json:"id"`
//		Name string `json:"name"`
//	}
//
//	user := User{ID: 1, Name: "John"}
//
//	// Convert to JSON
//	jsonData, err := helpers.ToJSON(user)
//
//	// Convert from JSON (returns pointer)
//	userPtr, err := helpers.FromJSON[User](jsonData)
//
//	// Convert from JSON (returns value)
//	userVal, err := helpers.FromJSONValue[User](jsonData)
//
//	// Pretty print
//	prettyJSON, err := helpers.PrettyPrint(user)
package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ToJSON converts any value to JSON bytes using Go generics
func ToJSON[T any](data T) ([]byte, error) {
	return json.Marshal(data)
}

// FromJSON converts JSON bytes to a struct and returns a pointer to the result
func FromJSON[T any](jsonData []byte) (*T, error) {
	var result T
	err := json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &result, nil
}

// FromJSONValue converts JSON bytes to a struct and returns the value (not pointer)
func FromJSONValue[T any](jsonData []byte) (T, error) {
	var result T
	err := json.Unmarshal(jsonData, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return result, nil
}

// UnmarshalJSON unmarshals JSON bytes into the provided interface
// This is a compatibility function for working with interface{} types
func UnmarshalJSON(jsonData []byte, v interface{}) error {
	return json.Unmarshal(jsonData, v)
}

// FromReader reads JSON from an io.Reader and converts it to a struct
func FromReader[T any](reader io.Reader) (*T, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON data: %w", err)
	}
	return FromJSON[T](data)
}

// FromReaderValue reads JSON from an io.Reader and converts it to a struct value
func FromReaderValue[T any](reader io.Reader) (T, error) {
	var result T
	data, err := io.ReadAll(reader)
	if err != nil {
		return result, fmt.Errorf("failed to read JSON data: %w", err)
	}
	return FromJSONValue[T](data)
}

// ToReader converts a struct to a JSON reader
func ToReader[T any](data T) (io.Reader, error) {
	jsonData, err := ToJSON(data)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(jsonData), nil
}

// ToString converts a struct to a JSON string
func ToString[T any](data T) (string, error) {
	jsonData, err := ToJSON(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromString converts a JSON string to a struct
func FromString[T any](jsonStr string) (*T, error) {
	return FromJSON[T]([]byte(jsonStr))
}

// FromStringValue converts a JSON string to a struct value
func FromStringValue[T any](jsonStr string) (T, error) {
	return FromJSONValue[T]([]byte(jsonStr))
}

// PrettyPrint returns a formatted JSON string
func PrettyPrint[T any](data T) (string, error) {
	jsonData, err := ToJSON(data)
	if err != nil {
		return "", err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, jsonData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %w", err)
	}
	return prettyJSON.String(), nil
}

// PrettyPrintWithIndent returns a formatted JSON string with custom indentation
func PrettyPrintWithIndent[T any](data T, prefix, indent string) (string, error) {
	jsonData, err := ToJSON(data)
	if err != nil {
		return "", err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, jsonData, prefix, indent)
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %w", err)
	}
	return prettyJSON.String(), nil
}

// ValidateJSON checks if a byte array contains valid JSON
func ValidateJSON(jsonData []byte) bool {
	return json.Valid(jsonData)
}

// ValidateJSONString checks if a string contains valid JSON
func ValidateJSONString(jsonStr string) bool {
	return json.Valid([]byte(jsonStr))
}

// CompactJSON removes whitespace from JSON bytes
func CompactJSON(jsonData []byte) ([]byte, error) {
	var compacted bytes.Buffer
	err := json.Compact(&compacted, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to compact JSON: %w", err)
	}
	return compacted.Bytes(), nil
}

// CompactJSONString removes whitespace from a JSON string
func CompactJSONString(jsonStr string) (string, error) {
	compacted, err := CompactJSON([]byte(jsonStr))
	if err != nil {
		return "", err
	}
	return string(compacted), nil
}

// MustToJSON converts a struct to JSON bytes, panics on error (use carefully)
func MustToJSON[T any](data T) []byte {
	result, err := ToJSON(data)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return result
}

// MustFromJSON converts JSON bytes to a struct, panics on error (use carefully)
func MustFromJSON[T any](jsonData []byte) *T {
	result, err := FromJSON[T](jsonData)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal JSON: %v", err))
	}
	return result
}

// MustPrettyPrint returns a formatted JSON string, panics on error (use carefully)
func MustPrettyPrint[T any](data T) string {
	result, err := PrettyPrint(data)
	if err != nil {
		panic(fmt.Sprintf("failed to pretty print JSON: %v", err))
	}
	return result
}

// IsEmptyJSON checks if JSON represents an empty object or array
func IsEmptyJSON(jsonData []byte) bool {
	if !ValidateJSON(jsonData) {
		return false
	}

	trimmed := strings.TrimSpace(string(jsonData))
	return trimmed == "{}" || trimmed == "[]" || trimmed == "null"
}

// MergeJSON merges multiple JSON objects into one (later objects override earlier ones)
func MergeJSON(jsonObjects ...[]byte) ([]byte, error) {
	if len(jsonObjects) == 0 {
		return []byte("{}"), nil
	}

	result := make(map[string]interface{})

	for _, jsonData := range jsonObjects {
		if !ValidateJSON(jsonData) {
			return nil, fmt.Errorf("invalid JSON data")
		}

		var obj map[string]interface{}
		if err := json.Unmarshal(jsonData, &obj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON object: %w", err)
		}

		// Merge the object into result
		for k, v := range obj {
			result[k] = v
		}
	}

	return json.Marshal(result)
}
