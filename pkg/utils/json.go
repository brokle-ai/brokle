package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// JSONMarshal marshals data to JSON with pretty formatting
func JSONMarshal(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// JSONUnmarshal unmarshals JSON data into a destination
func JSONUnmarshal(data []byte, dest interface{}) error {
	return json.Unmarshal(data, dest)
}

// JSONMarshalString marshals data to a JSON string
func JSONMarshalString(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// JSONUnmarshalString unmarshals a JSON string into a destination
func JSONUnmarshalString(jsonStr string, dest interface{}) error {
	return json.Unmarshal([]byte(jsonStr), dest)
}

// JSONPrettyPrint returns a pretty-printed JSON string
func JSONPrettyPrint(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// JSONMinify removes whitespace from JSON string
func JSONMinify(jsonStr string) (string, error) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	minified, err := json.Marshal(jsonData)
	if err != nil {
		return "", err
	}

	return string(minified), nil
}

// JSONValidate checks if a string is valid JSON
func JSONValidate(jsonStr string) bool {
	var jsonData interface{}
	return json.Unmarshal([]byte(jsonStr), &jsonData) == nil
}

// JSONGetValue extracts a value from JSON using dot notation path
func JSONGetValue(jsonData interface{}, path string) (interface{}, error) {
	if path == "" {
		return jsonData, nil
	}

	parts := strings.Split(path, ".")
	current := jsonData

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, exists := v[part]; exists {
				current = val
			} else {
				return nil, fmt.Errorf("path not found: %s", path)
			}
		case []interface{}:
			index, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", part)
			}
			if index < 0 || index >= len(v) {
				return nil, fmt.Errorf("array index out of bounds: %d", index)
			}
			current = v[index]
		default:
			return nil, fmt.Errorf("cannot navigate through non-object/array at path: %s", path)
		}
	}

	return current, nil
}

// JSONSetValue sets a value in JSON using dot notation path
func JSONSetValue(jsonData map[string]interface{}, path string, value interface{}) error {
	if path == "" {
		return errors.New("empty path")
	}

	parts := strings.Split(path, ".")
	current := jsonData

	// Navigate to the parent of the target
	for i := range len(parts) - 1 {
		part := parts[i]

		if next, exists := current[part]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return fmt.Errorf("path conflict: %s is not an object", part)
			}
		} else {
			// Create new object
			newObj := make(map[string]interface{})
			current[part] = newObj
			current = newObj
		}
	}

	// Set the final value
	current[parts[len(parts)-1]] = value
	return nil
}

// JSONMerge merges multiple JSON objects into one
func JSONMerge(jsons ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, jsonObj := range jsons {
		for key, value := range jsonObj {
			result[key] = value
		}
	}

	return result
}

// JSONDeepMerge deeply merges multiple JSON objects
func JSONDeepMerge(jsons ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, jsonObj := range jsons {
		deepMergeInto(result, jsonObj)
	}

	return result
}

func deepMergeInto(dest, src map[string]interface{}) {
	for key, srcValue := range src {
		if destValue, exists := dest[key]; exists {
			// Both exist, check if they're maps
			if destMap, ok := destValue.(map[string]interface{}); ok {
				if srcMap, ok := srcValue.(map[string]interface{}); ok {
					// Both are maps, merge recursively
					deepMergeInto(destMap, srcMap)
					continue
				}
			}
		}
		// Either destination doesn't exist or types don't match, overwrite
		dest[key] = srcValue
	}
}

// JSONClone creates a deep copy of JSON data
func JSONClone(data interface{}) (interface{}, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var clone interface{}
	if err := json.Unmarshal(jsonBytes, &clone); err != nil {
		return nil, err
	}

	return clone, nil
}

// JSONDiff finds differences between two JSON objects
func JSONDiff(obj1, obj2 map[string]interface{}) map[string]interface{} {
	diff := make(map[string]interface{})

	// Find keys that are different or missing in obj2
	for key, val1 := range obj1 {
		if val2, exists := obj2[key]; exists {
			if !reflect.DeepEqual(val1, val2) {
				diff[key] = map[string]interface{}{
					"old": val1,
					"new": val2,
				}
			}
		} else {
			diff[key] = map[string]interface{}{
				"old":     val1,
				"deleted": true,
			}
		}
	}

	// Find keys that are new in obj2
	for key, val2 := range obj2 {
		if _, exists := obj1[key]; !exists {
			diff[key] = map[string]interface{}{
				"new":   val2,
				"added": true,
			}
		}
	}

	return diff
}

// JSONFlatten flattens a nested JSON object using dot notation
func JSONFlatten(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	flattenRecursive("", data, result)
	return result
}

func flattenRecursive(prefix string, data map[string]interface{}, result map[string]interface{}) {
	for key, value := range data {
		newKey := key
		if prefix != "" {
			newKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			flattenRecursive(newKey, v, result)
		case []interface{}:
			for i, item := range v {
				arrayKey := newKey + "[" + strconv.Itoa(i) + "]"
				if itemMap, ok := item.(map[string]interface{}); ok {
					flattenRecursive(arrayKey, itemMap, result)
				} else {
					result[arrayKey] = item
				}
			}
		default:
			result[newKey] = value
		}
	}
}

// JSONUnflatten converts a flattened JSON object back to nested structure
func JSONUnflatten(flatData map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range flatData {
		setNestedValue(result, key, value)
	}

	return result
}

func setNestedValue(data map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := data

	for i := range len(parts) - 1 {
		part := parts[i]

		// Handle array notation
		if strings.Contains(part, "[") {
			// This is a more complex case that would require array handling
			// For simplicity, treating as regular key for now
		}

		if _, exists := current[part]; !exists {
			current[part] = make(map[string]interface{})
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		}
	}

	current[parts[len(parts)-1]] = value
}

// JSONToQueryString converts a JSON object to URL query string
func JSONToQueryString(data map[string]interface{}) string {
	var parts []string

	for key, value := range data {
		if value != nil {
			valueStr := fmt.Sprintf("%v", value)
			parts = append(parts, key+"="+valueStr)
		}
	}

	return strings.Join(parts, "&")
}

// JSONFromReader reads and parses JSON from an io.Reader
func JSONFromReader(reader io.Reader, dest interface{}) error {
	decoder := json.NewDecoder(reader)
	return decoder.Decode(dest)
}

// JSONToBuffer writes JSON data to a buffer
func JSONToBuffer(data interface{}) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)

	if err := encoder.Encode(data); err != nil {
		return nil, err
	}

	return buffer, nil
}

// JSONStreamDecode decodes multiple JSON objects from a stream
func JSONStreamDecode(reader io.Reader, callback func(interface{}) error) error {
	decoder := json.NewDecoder(reader)

	for {
		var data interface{}
		if err := decoder.Decode(&data); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := callback(data); err != nil {
			return err
		}
	}

	return nil
}

// JSONExtractKeys extracts all keys from a JSON object recursively
func JSONExtractKeys(data interface{}) []string {
	var keys []string
	extractKeysRecursive(data, "", &keys)
	return keys
}

func extractKeysRecursive(data interface{}, prefix string, keys *[]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}
			*keys = append(*keys, fullKey)
			extractKeysRecursive(value, fullKey, keys)
		}
	case []interface{}:
		for i, item := range v {
			arrayKey := prefix + "[" + strconv.Itoa(i) + "]"
			extractKeysRecursive(item, arrayKey, keys)
		}
	}
}

// JSONSize calculates the approximate size of JSON data in bytes
func JSONSize(data interface{}) (int, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}
	return len(jsonBytes), nil
}
