package template

import (
	"bytes"
	"text/template"

	"github.com/dwrtz/sink/internal/processor"
)

type Engine struct {
	templateText string
}

func NewEngine(templateText string) *Engine {
	return &Engine{templateText: templateText}
}

func (e *Engine) Execute(files []processor.FileInfo) (string, error) {
	tmpl, err := template.New("markdown").Parse(e.templateText)
	if err != nil {
		return "", err
	}

	data := struct {
		Files []processor.FileInfo
	}{
		Files: files,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
