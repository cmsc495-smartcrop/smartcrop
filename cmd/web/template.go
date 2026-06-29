package main

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/cmsc495-smartcrop/smartcrop/ui"

	"github.com/labstack/echo/v5"
)

var funcMap = template.FuncMap{
	"currentYear": func() int { return time.Now().Year() },
}

type Renderer struct {
	templates map[string]*template.Template
}

func (r *Renderer) Render(c *echo.Context, w io.Writer, name string, data any) error {
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %q not found", name)
	}
	if strings.HasPrefix(name, "partials/") {
		stem := strings.TrimSuffix(filepath.Base(name), ".gohtml")
		return t.ExecuteTemplate(w, stem, data)
	}
	return t.Execute(w, data)
}

func NewTemplateCache() (*Renderer, error) {
	cache := map[string]*template.Template{}

	viewFS := ui.ViewsFS()

	pages, err := fs.Glob(viewFS, "pages/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		fileName := filepath.Base(page)
		mapKey := "pages/" + fileName

		patterns := []string{
			"layouts/base.gohtml",
			"partials/*.gohtml",
			page,
		}

		ts, err := template.New(fileName).Funcs(funcMap).ParseFS(viewFS, patterns...)
		if err != nil {
			return nil, err
		}

		cache[mapKey] = ts
	}

	partials, err := fs.Glob(viewFS, "partials/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, partial := range partials {
		fileName := filepath.Base(partial)
		mapKey := "partials/" + fileName
		ts, err := template.New(fileName).Funcs(funcMap).ParseFS(viewFS, "partials/*.gohtml")
		if err != nil {
			return nil, err
		}
		cache[mapKey] = ts
	}

	return &Renderer{templates: cache}, nil
}
