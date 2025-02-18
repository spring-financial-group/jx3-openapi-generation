package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// FirstCharToLower converts the first character of a string to lowercase
func FirstCharToLower(s string) string {
	if len(s) == 0 {
		return s
	}
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

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

// getMajorVersion returns the major version of the package
// if major version is >1, it will return v{majorVersion}
func GetMajorVersion(rawVersion string) string {
	// get the version string with regex and split it by the dot
	regex := regexp.MustCompile(`v(\d+\.\d+\.\d+)`)
	if !regex.MatchString(rawVersion) {
		return ""
	}
	subMatch := regex.FindStringSubmatch(rawVersion)
	if len(subMatch) < 2 {
		return ""
	}
	version := strings.Split(subMatch[1], ".")
	if len(version) < 1 {
		return ""
	}
	majorVersion := version[0]
	if majorVersion == "0" || majorVersion == "1" {
		return ""
	}
	return fmt.Sprintf("v%s", majorVersion)
}
