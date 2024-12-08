package linenumbers

import (
	"fmt"
	"strings"
)

func AddLineNumbers(content string) string {
	lines := strings.Split(content, "\n")
	width := len(fmt.Sprint(len(lines)))
	format := fmt.Sprintf("%%%dd | %%s", width)

	var result strings.Builder
	for i, line := range lines {
		result.WriteString(fmt.Sprintf(format, i+1, line))
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}
