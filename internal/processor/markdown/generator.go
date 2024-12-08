package markdown

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dwrtz/sink/internal/processor"
	"github.com/dwrtz/sink/internal/processor/comments"
	"github.com/dwrtz/sink/internal/processor/linenumbers"
)

type Config struct {
	NoCodeBlock   bool
	LineNumbers   bool
	StripComments bool
}

type Generator struct {
	config Config
}

func NewGenerator(config Config) *Generator {
	return &Generator{config: config}
}

func (g *Generator) Generate(files []processor.FileInfo) (string, error) {
	var content strings.Builder

	// Generate table of contents
	content.WriteString("# Table of Contents\n")
	for _, file := range files {
		content.WriteString(fmt.Sprintf("- %s\n", file.Path))
	}
	content.WriteString("\n")

	// Generate content for each file
	for _, file := range files {
		content.WriteString(g.generateFileSection(file))
	}

	return content.String(), nil
}

func (g *Generator) generateFileSection(file processor.FileInfo) string {
	var section strings.Builder

	// File header
	section.WriteString(fmt.Sprintf("## File: %s\n\n", file.Path))
	section.WriteString(fmt.Sprintf("- Extension: %s\n", filepath.Ext(file.Path)))
	section.WriteString(fmt.Sprintf("- Language: %s\n", file.Language))
	section.WriteString(fmt.Sprintf("- Size: %d bytes\n", file.Size))
	section.WriteString(fmt.Sprintf("- Created: %s\n", file.Created.Format("2006-01-02 15:04:05")))
	section.WriteString(fmt.Sprintf("- Modified: %s\n\n", file.Modified.Format("2006-01-02 15:04:05")))

	// Code content
	section.WriteString("### Code\n\n")

	content := file.Content
	if g.config.StripComments {
		content = comments.StripComments(content, file.Language)
	}
	if g.config.LineNumbers {
		content = linenumbers.AddLineNumbers(content)
	}

	if !g.config.NoCodeBlock {
		section.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", file.Language, content))
	} else {
		section.WriteString(fmt.Sprintf("%s\n\n", content))
	}

	return section.String()
}
