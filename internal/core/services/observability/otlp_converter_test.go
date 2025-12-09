package observability

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test extractGenericInput with LLM messages (highest priority)
func TestExtractGenericInput_GenAIMessages_Array(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.input.messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "Hello"},
		},
	}

	value, mimeType := extractGenericInput(attrs)

	assert.Contains(t, value, `"role":"user"`)
	assert.Contains(t, value, `"content":"Hello"`)
	assert.Equal(t, "application/json", mimeType)
}

func TestExtractGenericInput_GenAIMessages_String(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.input.messages": `[{"role":"user","content":"Hello"}]`,
	}

	value, mimeType := extractGenericInput(attrs)

	assert.Equal(t, `[{"role":"user","content":"Hello"}]`, value)
	assert.Equal(t, "application/json", mimeType)
}

// Test extractGenericInput with input.value (fallback)
func TestExtractGenericInput_InputValue_String(t *testing.T) {
	attrs := map[string]interface{}{
		"input.value": `{"location":"Bangalore","units":"celsius"}`,
	}

	value, mimeType := extractGenericInput(attrs)

	assert.Equal(t, `{"location":"Bangalore","units":"celsius"}`, value)
	assert.Equal(t, "application/json", mimeType) // Auto-detected
}

func TestExtractGenericInput_InputValue_StringWithMimeType(t *testing.T) {
	attrs := map[string]interface{}{
		"input.value":     "Hello, World!",
		"input.mime_type": "text/plain",
	}

	value, mimeType := extractGenericInput(attrs)

	assert.Equal(t, "Hello, World!", value)
	assert.Equal(t, "text/plain", mimeType)
}

// Test extractGenericInput with object input (new capability)
func TestExtractGenericInput_InputValue_Object(t *testing.T) {
	attrs := map[string]interface{}{
		"input.value": map[string]interface{}{
			"location": "Bangalore",
			"units":    "celsius",
		},
	}

	value, mimeType := extractGenericInput(attrs)

	assert.Contains(t, value, `"location":"Bangalore"`)
	assert.Contains(t, value, `"units":"celsius"`)
	assert.Equal(t, "application/json", mimeType)
}

// Test extractGenericInput with array input (new capability)
func TestExtractGenericInput_InputValue_Array(t *testing.T) {
	attrs := map[string]interface{}{
		"input.value": []interface{}{"arg1", "arg2", 123},
	}

	value, mimeType := extractGenericInput(attrs)

	assert.Equal(t, `["arg1","arg2",123]`, value)
	assert.Equal(t, "application/json", mimeType)
}

// Test priority: gen_ai.input.messages takes precedence over input.value
func TestExtractGenericInput_PriorityOrder(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.input.messages": `[{"role":"user","content":"high priority"}]`,
		"input.value":           `{"generic":"low priority"}`,
	}

	value, mimeType := extractGenericInput(attrs)

	assert.Equal(t, `[{"role":"user","content":"high priority"}]`, value)
	assert.Equal(t, "application/json", mimeType)
}

// Test extractGenericInput returns empty when no input
func TestExtractGenericInput_Empty(t *testing.T) {
	attrs := map[string]interface{}{
		"other.attribute": "value",
	}

	value, mimeType := extractGenericInput(attrs)

	assert.Equal(t, "", value)
	assert.Equal(t, "", mimeType)
}

// Test extractGenericOutput with LLM messages (highest priority)
func TestExtractGenericOutput_GenAIMessages_Array(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.output.messages": []interface{}{
			map[string]interface{}{"role": "assistant", "content": "Hello back!"},
		},
	}

	value, mimeType := extractGenericOutput(attrs)

	assert.Contains(t, value, `"role":"assistant"`)
	assert.Contains(t, value, `"content":"Hello back!"`)
	assert.Equal(t, "application/json", mimeType)
}

func TestExtractGenericOutput_GenAIMessages_String(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.output.messages": `[{"role":"assistant","content":"Response"}]`,
	}

	value, mimeType := extractGenericOutput(attrs)

	assert.Equal(t, `[{"role":"assistant","content":"Response"}]`, value)
	assert.Equal(t, "application/json", mimeType)
}

// Test extractGenericOutput with output.value (fallback)
func TestExtractGenericOutput_OutputValue_String(t *testing.T) {
	attrs := map[string]interface{}{
		"output.value": `{"temperature":25,"conditions":"sunny"}`,
	}

	value, mimeType := extractGenericOutput(attrs)

	assert.Equal(t, `{"temperature":25,"conditions":"sunny"}`, value)
	assert.Equal(t, "application/json", mimeType)
}

// Test extractGenericOutput with object output (new capability)
func TestExtractGenericOutput_OutputValue_Object(t *testing.T) {
	attrs := map[string]interface{}{
		"output.value": map[string]interface{}{
			"temperature": 25,
			"conditions":  "sunny",
		},
	}

	value, mimeType := extractGenericOutput(attrs)

	assert.Contains(t, value, `"temperature":25`)
	assert.Contains(t, value, `"conditions":"sunny"`)
	assert.Equal(t, "application/json", mimeType)
}

// Test extractGenericOutput with array output (new capability)
func TestExtractGenericOutput_OutputValue_Array(t *testing.T) {
	attrs := map[string]interface{}{
		"output.value": []interface{}{"result1", "result2"},
	}

	value, mimeType := extractGenericOutput(attrs)

	assert.Equal(t, `["result1","result2"]`, value)
	assert.Equal(t, "application/json", mimeType)
}

// Test priority: gen_ai.output.messages takes precedence
func TestExtractGenericOutput_PriorityOrder(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.output.messages": `[{"role":"assistant","content":"high priority"}]`,
		"output.value":           `{"generic":"low priority"}`,
	}

	value, mimeType := extractGenericOutput(attrs)

	assert.Equal(t, `[{"role":"assistant","content":"high priority"}]`, value)
	assert.Equal(t, "application/json", mimeType)
}

// Test extractToolMetadata
func TestExtractToolMetadata_ToolName(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.tool.name": "get_weather",
	}
	payload := make(map[string]interface{})

	extractToolMetadata(attrs, payload)

	assert.Equal(t, "get_weather", payload["tool_name"])
}

func TestExtractToolMetadata_ToolParameters_String(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.tool.name":       "get_weather",
		"gen_ai.tool.parameters": `{"location":"Bangalore"}`,
	}
	payload := make(map[string]interface{})

	extractToolMetadata(attrs, payload)

	assert.Equal(t, "get_weather", payload["tool_name"])
	assert.Equal(t, `{"location":"Bangalore"}`, payload["tool_parameters"])
}

func TestExtractToolMetadata_ToolParameters_Object(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.tool.name": "get_weather",
		"gen_ai.tool.parameters": map[string]interface{}{
			"location": "Bangalore",
			"units":    "celsius",
		},
	}
	payload := make(map[string]interface{})

	extractToolMetadata(attrs, payload)

	assert.Equal(t, "get_weather", payload["tool_name"])
	assert.Contains(t, payload["tool_parameters"].(string), `"location":"Bangalore"`)
}

func TestExtractToolMetadata_ToolResult_String(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.tool.name":   "get_weather",
		"gen_ai.tool.result": `{"temp":25}`,
	}
	payload := make(map[string]interface{})

	extractToolMetadata(attrs, payload)

	assert.Equal(t, "get_weather", payload["tool_name"])
	assert.Equal(t, `{"temp":25}`, payload["tool_result"])
}

func TestExtractToolMetadata_ToolResult_Object(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.tool.name": "get_weather",
		"gen_ai.tool.result": map[string]interface{}{
			"temp":       25,
			"conditions": "sunny",
		},
	}
	payload := make(map[string]interface{})

	extractToolMetadata(attrs, payload)

	assert.Equal(t, "get_weather", payload["tool_name"])
	assert.Contains(t, payload["tool_result"].(string), `"temp":25`)
}

func TestExtractToolMetadata_ToolCallID(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.tool.name":    "get_weather",
		"gen_ai.tool.call.id": "call_abc123",
	}
	payload := make(map[string]interface{})

	extractToolMetadata(attrs, payload)

	assert.Equal(t, "get_weather", payload["tool_name"])
	assert.Equal(t, "call_abc123", payload["tool_call_id"])
}

func TestExtractToolMetadata_AllFields(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.tool.name":       "search_database",
		"gen_ai.tool.parameters": `{"query":"SELECT * FROM users"}`,
		"gen_ai.tool.result":     `{"rows":10}`,
		"gen_ai.tool.call.id":    "call_xyz789",
	}
	payload := make(map[string]interface{})

	extractToolMetadata(attrs, payload)

	assert.Equal(t, "search_database", payload["tool_name"])
	assert.Equal(t, `{"query":"SELECT * FROM users"}`, payload["tool_parameters"])
	assert.Equal(t, `{"rows":10}`, payload["tool_result"])
	assert.Equal(t, "call_xyz789", payload["tool_call_id"])
}

func TestExtractToolMetadata_Empty(t *testing.T) {
	attrs := map[string]interface{}{
		"other.attribute": "value",
	}
	payload := make(map[string]interface{})

	extractToolMetadata(attrs, payload)

	assert.Nil(t, payload["tool_name"])
	assert.Nil(t, payload["tool_parameters"])
	assert.Nil(t, payload["tool_result"])
	assert.Nil(t, payload["tool_call_id"])
}

// Test validateMimeType
func TestValidateMimeType_AutoDetectJSON(t *testing.T) {
	result := validateMimeType(`{"key":"value"}`, "")
	assert.Equal(t, "application/json", result)
}

func TestValidateMimeType_AutoDetectPlainText(t *testing.T) {
	result := validateMimeType("Hello World", "")
	assert.Equal(t, "text/plain", result)
}

func TestValidateMimeType_DeclaredValid(t *testing.T) {
	result := validateMimeType(`{"key":"value"}`, "application/json")
	assert.Equal(t, "application/json", result)
}

func TestValidateMimeType_DeclaredInvalid(t *testing.T) {
	// Declared JSON but content is not valid JSON
	result := validateMimeType("not valid json", "application/json")
	assert.Equal(t, "text/plain", result)
}

// Test truncateWithIndicator
func TestTruncateWithIndicator_NoTruncation(t *testing.T) {
	value := "short string"
	result, truncated := truncateWithIndicator(value, 100)

	assert.Equal(t, value, result)
	assert.False(t, truncated)
}

func TestTruncateWithIndicator_Truncation(t *testing.T) {
	value := "this is a longer string that needs truncation"
	result, truncated := truncateWithIndicator(value, 20)

	assert.True(t, truncated)
	assert.Contains(t, result, "...[truncated]")
	assert.True(t, len(result) < len(value)+15) // Original truncated + indicator
}

// Regression test: extractGenAIFields should NOT overwrite payload["input"]/["output"]
// These are already set by createSpanEvent with proper truncation and MIME type handling.
func TestExtractGenAIFields_DoesNotOverwriteExistingInputOutput(t *testing.T) {
	attrs := map[string]interface{}{
		"gen_ai.input.messages":  `[{"role":"user","content":"from attrs"}]`,
		"gen_ai.output.messages": `[{"role":"assistant","content":"from attrs"}]`,
		"gen_ai.provider.name":   "openai",
		"gen_ai.request.model":   "gpt-4",
	}

	payload := map[string]interface{}{
		"input":           "already set with truncation",
		"input_mime_type": "application/json",
		"output":          "already set with truncation",
	}

	extractGenAIFields(attrs, payload)

	// Input/output should NOT be overwritten by extractGenAIFields
	assert.Equal(t, "already set with truncation", payload["input"])
	assert.Equal(t, "already set with truncation", payload["output"])
	assert.Equal(t, "application/json", payload["input_mime_type"])

	// Other fields should still be extracted
	assert.Equal(t, "openai", payload["provider"])
	assert.Equal(t, "gpt-4", payload["model_name"])
}
