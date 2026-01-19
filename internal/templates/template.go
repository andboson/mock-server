package templates

import (
	"embed"
	"fmt"
	"html/template"
)

//go:embed *.tmpl
var embedTpl embed.FS

// Templates holds parsed a templates data
type Templates struct {
	Tpls *template.Template
}

// NewTemplates returns an instance of Templates
func NewTemplates() (*Templates, error) {
	path := "index.tmpl"
	t, err := template.New(path).ParseFS(embedTpl, path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Templates{
		Tpls: t,
	}, nil
}
