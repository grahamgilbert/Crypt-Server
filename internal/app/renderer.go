package app

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	baseLayout string
	pageDir    string
	cache      map[string]*template.Template
}

func NewRenderer(baseLayout, pageDir string) *Renderer {
	return &Renderer{
		baseLayout: baseLayout,
		pageDir:    pageDir,
		cache:      make(map[string]*template.Template),
	}
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data any) error {
	page, ok := r.cache[name]
	if !ok {
		layout := r.baseLayout
		pagePath := filepath.Join(r.pageDir, name+".html")
		parsed, err := template.New("base").Funcs(template.FuncMap{
			"add": func(a, b int) int { return a + b },
			"sub": func(a, b int) int { return a - b },
		}).ParseFiles(layout, pagePath)
		if err != nil {
			return fmt.Errorf("parse templates: %w", err)
		}
		r.cache[name] = parsed
		page = parsed
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := page.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	return nil
}
