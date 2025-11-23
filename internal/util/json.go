package util

import (
	"bytes"
	"encoding/json"
)

func IsJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func PrettyJSON(data []byte) (string, error) {
	// First, unmarshal to decode Unicode escape sequences (e.g., \u00f3 -> ó)
	var js interface{}
	if err := json.Unmarshal(data, &js); err != nil {
		return "", err
	}

	// Then marshal back with indentation, which will properly encode Unicode characters
	// without unnecessary escaping for common characters
	out := bytes.Buffer{}
	encoder := json.NewEncoder(&out)
	encoder.SetIndent("", "    ")
	encoder.SetEscapeHTML(false) // Don't escape HTML characters like <, >, &
	if err := encoder.Encode(js); err != nil {
		return "", err
	}

	// Remove trailing newline added by Encode
	result := out.String()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}

func ParseJSON(text string) (map[string]any, error) {
	var js map[string]any
	if err := json.Unmarshal([]byte(text), &js); err != nil {
		return nil, err
	}
	return js, nil
}

func EncodeJSON(data any) ([]byte, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return b, nil
}
