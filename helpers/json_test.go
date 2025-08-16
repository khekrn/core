package helpers

import (
	"strings"
	"testing"
)

type TestStruct struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age,omitempty"`
}

func TestToJSON(t *testing.T) {
	data := TestStruct{ID: 1, Name: "John", Age: 30}

	jsonData, err := ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if !ValidateJSON(jsonData) {
		t.Error("Generated JSON is not valid")
	}

	// Should contain the expected fields
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, `"id":1`) {
		t.Error("JSON should contain id field")
	}
	if !strings.Contains(jsonStr, `"name":"John"`) {
		t.Error("JSON should contain name field")
	}
}

func TestFromJSON(t *testing.T) {
	jsonData := []byte(`{"id":1,"name":"John","age":30}`)

	result, err := FromJSON[TestStruct](jsonData)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
	if result.Name != "John" {
		t.Errorf("Expected name John, got %s", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("Expected age 30, got %d", result.Age)
	}
}

func TestFromJSONValue(t *testing.T) {
	jsonData := []byte(`{"id":1,"name":"John","age":30}`)

	result, err := FromJSONValue[TestStruct](jsonData)
	if err != nil {
		t.Fatalf("FromJSONValue failed: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
	if result.Name != "John" {
		t.Errorf("Expected name John, got %s", result.Name)
	}
}

func TestFromReader(t *testing.T) {
	jsonStr := `{"id":1,"name":"John","age":30}`
	reader := strings.NewReader(jsonStr)

	result, err := FromReader[TestStruct](reader)
	if err != nil {
		t.Fatalf("FromReader failed: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

func TestFromReaderValue(t *testing.T) {
	jsonStr := `{"id":1,"name":"John","age":30}`
	reader := strings.NewReader(jsonStr)

	result, err := FromReaderValue[TestStruct](reader)
	if err != nil {
		t.Fatalf("FromReaderValue failed: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

func TestToReader(t *testing.T) {
	data := TestStruct{ID: 1, Name: "John", Age: 30}

	reader, err := ToReader(data)
	if err != nil {
		t.Fatalf("ToReader failed: %v", err)
	}

	// Read back and verify
	result, err := FromReader[TestStruct](reader)
	if err != nil {
		t.Fatalf("Failed to read back from reader: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

func TestToString(t *testing.T) {
	data := TestStruct{ID: 1, Name: "John"}

	jsonStr, err := ToString(data)
	if err != nil {
		t.Fatalf("ToString failed: %v", err)
	}

	if !ValidateJSONString(jsonStr) {
		t.Error("Generated JSON string is not valid")
	}
}

func TestFromString(t *testing.T) {
	jsonStr := `{"id":1,"name":"John","age":30}`

	result, err := FromString[TestStruct](jsonStr)
	if err != nil {
		t.Fatalf("FromString failed: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

func TestFromStringValue(t *testing.T) {
	jsonStr := `{"id":1,"name":"John","age":30}`

	result, err := FromStringValue[TestStruct](jsonStr)
	if err != nil {
		t.Fatalf("FromStringValue failed: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

func TestPrettyPrint(t *testing.T) {
	data := TestStruct{ID: 1, Name: "John", Age: 30}

	prettyJSON, err := PrettyPrint(data)
	if err != nil {
		t.Fatalf("PrettyPrint failed: %v", err)
	}

	// Should contain newlines and indentation
	if !strings.Contains(prettyJSON, "\n") {
		t.Error("Pretty printed JSON should contain newlines")
	}
	if !strings.Contains(prettyJSON, "  ") {
		t.Error("Pretty printed JSON should contain indentation")
	}
}

func TestPrettyPrintWithIndent(t *testing.T) {
	data := TestStruct{ID: 1, Name: "John"}

	prettyJSON, err := PrettyPrintWithIndent(data, "", "\t")
	if err != nil {
		t.Fatalf("PrettyPrintWithIndent failed: %v", err)
	}

	// Should contain tabs
	if !strings.Contains(prettyJSON, "\t") {
		t.Error("Pretty printed JSON should contain tab indentation")
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData []byte
		valid    bool
	}{
		{"Valid object", []byte(`{"id":1,"name":"John"}`), true},
		{"Valid array", []byte(`[1,2,3]`), true},
		{"Valid null", []byte(`null`), true},
		{"Valid string", []byte(`"hello"`), true},
		{"Valid number", []byte(`123`), true},
		{"Invalid JSON", []byte(`{"id":1,"name":}`), false},
		{"Empty", []byte(``), false},
		{"Invalid syntax", []byte(`{id:1}`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateJSON(tt.jsonData)
			if result != tt.valid {
				t.Errorf("ValidateJSON(%s) = %v, want %v", string(tt.jsonData), result, tt.valid)
			}
		})
	}
}

func TestValidateJSONString(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		valid   bool
	}{
		{"Valid object", `{"id":1,"name":"John"}`, true},
		{"Valid array", `[1,2,3]`, true},
		{"Invalid JSON", `{"id":1,"name":}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateJSONString(tt.jsonStr)
			if result != tt.valid {
				t.Errorf("ValidateJSONString(%s) = %v, want %v", tt.jsonStr, result, tt.valid)
			}
		})
	}
}

func TestCompactJSON(t *testing.T) {
	jsonData := []byte(`{
		"id": 1,
		"name": "John",
		"age": 30
	}`)

	compacted, err := CompactJSON(jsonData)
	if err != nil {
		t.Fatalf("CompactJSON failed: %v", err)
	}

	compactedStr := string(compacted)
	if strings.Contains(compactedStr, "\n") || strings.Contains(compactedStr, "\t") || strings.Contains(compactedStr, "  ") {
		t.Error("Compacted JSON should not contain whitespace")
	}

	// Should still be valid JSON
	if !ValidateJSON(compacted) {
		t.Error("Compacted JSON should be valid")
	}
}

func TestCompactJSONString(t *testing.T) {
	jsonStr := `{
		"id": 1,
		"name": "John"
	}`

	compacted, err := CompactJSONString(jsonStr)
	if err != nil {
		t.Fatalf("CompactJSONString failed: %v", err)
	}

	if strings.Contains(compacted, "\n") || strings.Contains(compacted, "\t") {
		t.Error("Compacted JSON string should not contain whitespace")
	}
}

func TestMustToJSON(t *testing.T) {
	data := TestStruct{ID: 1, Name: "John"}

	// Should not panic for valid data
	jsonData := MustToJSON(data)
	if !ValidateJSON(jsonData) {
		t.Error("MustToJSON should produce valid JSON")
	}
}

func TestMustFromJSON(t *testing.T) {
	jsonData := []byte(`{"id":1,"name":"John"}`)

	// Should not panic for valid JSON
	result := MustFromJSON[TestStruct](jsonData)
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

func TestMustPrettyPrint(t *testing.T) {
	data := TestStruct{ID: 1, Name: "John"}

	// Should not panic for valid data
	prettyJSON := MustPrettyPrint(data)
	if !strings.Contains(prettyJSON, "\n") {
		t.Error("MustPrettyPrint should produce formatted JSON")
	}
}

func TestIsEmptyJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData []byte
		isEmpty  bool
	}{
		{"Empty object", []byte(`{}`), true},
		{"Empty array", []byte(`[]`), true},
		{"Null", []byte(`null`), true},
		{"Empty with spaces", []byte(`  {}  `), true},
		{"Non-empty object", []byte(`{"id":1}`), false},
		{"Non-empty array", []byte(`[1]`), false},
		{"Invalid JSON", []byte(`{`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmptyJSON(tt.jsonData)
			if result != tt.isEmpty {
				t.Errorf("IsEmptyJSON(%s) = %v, want %v", string(tt.jsonData), result, tt.isEmpty)
			}
		})
	}
}

func TestMergeJSON(t *testing.T) {
	json1 := []byte(`{"a":1,"b":2}`)
	json2 := []byte(`{"b":3,"c":4}`)
	json3 := []byte(`{"c":5,"d":6}`)

	merged, err := MergeJSON(json1, json2, json3)
	if err != nil {
		t.Fatalf("MergeJSON failed: %v", err)
	}

	// Parse the result to verify merging
	result, err := FromJSONValue[map[string]interface{}](merged)
	if err != nil {
		t.Fatalf("Failed to parse merged JSON: %v", err)
	}

	// Check that later values override earlier ones
	if result["a"] != 1.0 { // JSON numbers are float64
		t.Errorf("Expected a=1, got %v", result["a"])
	}
	if result["b"] != 3.0 { // Should be overridden by json2
		t.Errorf("Expected b=3, got %v", result["b"])
	}
	if result["c"] != 5.0 { // Should be overridden by json3
		t.Errorf("Expected c=5, got %v", result["c"])
	}
	if result["d"] != 6.0 {
		t.Errorf("Expected d=6, got %v", result["d"])
	}
}

func TestMergeJSON_EmptyInput(t *testing.T) {
	merged, err := MergeJSON()
	if err != nil {
		t.Fatalf("MergeJSON with no input failed: %v", err)
	}

	expected := `{}`
	if string(merged) != expected {
		t.Errorf("Expected %s, got %s", expected, string(merged))
	}
}

func TestMergeJSON_InvalidJSON(t *testing.T) {
	json1 := []byte(`{"a":1}`)
	json2 := []byte(`{invalid}`)

	_, err := MergeJSON(json1, json2)
	if err == nil {
		t.Error("MergeJSON should fail with invalid JSON")
	}
}

// Benchmark tests
func BenchmarkToJSON(b *testing.B) {
	data := TestStruct{ID: 1, Name: "John", Age: 30}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ToJSON(data)
	}
}

func BenchmarkFromJSON(b *testing.B) {
	jsonData := []byte(`{"id":1,"name":"John","age":30}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FromJSON[TestStruct](jsonData)
	}
}

func BenchmarkPrettyPrint(b *testing.B) {
	data := TestStruct{ID: 1, Name: "John", Age: 30}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = PrettyPrint(data)
	}
}
