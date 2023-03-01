package utils

import "encoding/json"

// NewPtr returns a pointer to the input
func NewPtr[T any](val T) *T {
	return &val
}

// MarshalJSON marshals the input to JSON with indentation & a trailing newline
func MarshalJSON[T any](dataStruct T) ([]byte, error) {
	marshall, err := json.MarshalIndent(dataStruct, "", "  ")
	if err != nil {
		return nil, err
	}
	marshall = append(marshall, "\n"...)
	return marshall, nil
}
