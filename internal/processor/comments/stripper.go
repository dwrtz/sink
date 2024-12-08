package comments

import (
	"regexp"
	"strings"
)

func StripComments(content, language string) string {
	switch language {
	case "go":
		return stripGoComments(content)
	case "python":
		return stripPythonComments(content)
	case "javascript":
		return stripJavaScriptComments(content)
	// Add more languages as needed
	default:
		return content
	}
}

func stripGoComments(content string) string {
	// Strip line comments
	lineComments := regexp.MustCompile(`//.*`)
	content = lineComments.ReplaceAllString(content, "")

	// Strip block comments
	blockComments := regexp.MustCompile(`(?s)/\*.*?\*/`)
	content = blockComments.ReplaceAllString(content, "")

	return strings.TrimSpace(content)
}

func stripPythonComments(content string) string {
	// Strip line comments
	lineComments := regexp.MustCompile(`#.*`)
	content = lineComments.ReplaceAllString(content, "")

	// Strip triple-quoted strings (docstrings)
	// Matches """...""" or '''...'''
	docStrings := regexp.MustCompile(`(?s)(\"\"\".*?\"\"\"|\'\'\'.*?\'\'\')`)
	content = docStrings.ReplaceAllString(content, "")

	return strings.TrimSpace(content)
}

func stripJavaScriptComments(content string) string {
	// Similar to Go comments
	lineComments := regexp.MustCompile(`//.*`)
	content = lineComments.ReplaceAllString(content, "")

	blockComments := regexp.MustCompile(`(?s)/\*.*?\*/`)
	content = blockComments.ReplaceAllString(content, "")

	return strings.TrimSpace(content)
}
