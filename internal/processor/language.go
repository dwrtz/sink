package processor

import "path/filepath"

// detectLanguage determines the programming language of a file based on its extension
func detectLanguage(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".java":
		return "java"
	case ".cpp", ".hpp", ".cc", ".hh":
		return "cpp"
	case ".c", ".h":
		return "c"
	// Add more language mappings as needed
	default:
		return "unknown"
	}
}
