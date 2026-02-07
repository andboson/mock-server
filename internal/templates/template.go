package templates

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
)

//go:embed *.tmpl
var embedTpl embed.FS

// Templates holds parsed template data
type Templates struct {
	Tpls *template.Template
}

// NewTemplates returns an instance of Templates
func NewTemplates() (*Templates, error) {
	funcMap := template.FuncMap{
		"deref": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"jsonMarshal": func(v interface{}) string {
			b, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Sprintf("error: %v", err)
			}
			return string(b)
		},
	}

	t, err := template.New("").Funcs(funcMap).ParseFS(embedTpl, "*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Templates{
		Tpls: t,
	}, nil
}
