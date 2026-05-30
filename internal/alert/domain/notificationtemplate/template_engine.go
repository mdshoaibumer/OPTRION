package notificationtemplate

import (
	"bytes"
	"text/template"
)

// TemplateEngine renders notification templates with variables and Markdown support.
type TemplateEngine struct{}

func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{}
}

func (te *TemplateEngine) Render(tmpl string, vars map[string]interface{}) (string, error) {
	t, err := template.New("notification").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, vars)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
