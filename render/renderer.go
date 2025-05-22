package render

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
)

// ViewSubdirectoryName is the standard subdirectory to be used when adding namespaced templates to the renderer.
const (
	ViewSubdirectoryName = "views"
	basepath             = "basepath"
)

// Renderer implements the necessary magic to render Hotwire Turbo frames and streams.
// Implements the `echo.Renderer` interface.
type Renderer struct {
	templates    NamespaceFS
	globalPaths  []string
	rootPage     string
	notFoundPage string
	globals      map[string]any
}

// New setups a new template renderer. rootPage is the Hotwire wrapping page that should be rendered,
func New(rootPage, rootPath string) (*Renderer, error) {
	r := &Renderer{
		templates: make(map[string]fs.FS),
		rootPage:  rootPage,
		globals:   template.FuncMap{},
	}
	basePath := rootPath
	if !strings.HasPrefix(basePath, "/") {
		basePath = fmt.Sprintf("/%s", basePath)
	}
	if basePath == "/" {
		basePath = ""
	}
	r.AddGlobal(basepath, basePath)

	return r, nil
}

func (r *Renderer) BasePath() string {
	basePath := r.globals[basepath].(string)
	if basePath == "" {
		return "/"
	}
	return basePath
}

func (r *Renderer) BasePathHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != r.BasePath() {
			r.Render(req.Context(), w, r.notFoundPage, req.URL.Path)
			return
		}

		err := r.Render(req.Context(), w, r.rootPage, nil)
		if err != nil {
			slog.Error("render base page", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (r *Renderer) AddNotFound(notfoundPath string) {
	r.notFoundPage = notfoundPath
}

// AddGlobal allows you to add a specific string replacement to the template renderer that will return
// a string only known at setup-time for a specific placeholder template, i.e. {{ global "basepath" }}
func (r *Renderer) AddGlobal(placeholder string, value any) {
	r.globals[placeholder] = value
}

// AddFS adds the templates from the passed filesystem to the template map.
// The key is the namespace, which is generally the name of the domain providing the views.
// If isGlobal is true, the namespace is added to the global paths and is accessible from
// any request for rendering. Global template files should have distinct names, as otherwise
// it can not be guaranteed which template is chosen.
func (r *Renderer) AddFS(namespace string, fsys fs.FS, isGlobal bool) error {
	err := r.templates.AddNamespace(namespace, ViewSubdirectoryName, fsys)
	if err != nil {
		return fmt.Errorf("add namespace filesystem %q", namespace)
	}
	if isGlobal {
		r.globalPaths = append(r.globalPaths, filepath.Join(namespace, "*.html"))
	}

	return nil
}

// Render finds the correct template in the template map based on name and namespace. It also automagically
// renders the correct template based on the passed template name and request or template name settings.
func (r *Renderer) Render(_ context.Context, w http.ResponseWriter, name string, data any) error {
	slog.Info("render", "page", name)
	tmpl, err := r.newTemplate(name)
	if err != nil {
		return fmt.Errorf("initialize template: %w", err)
	}

	err = tmpl.ExecuteTemplate(w, filepath.Base(name), data)
	if err != nil {
		return fmt.Errorf("execute render template %q", name)
	}

	return nil
}

func (r *Renderer) newTemplate(name string) (*template.Template, error) {
	tmpl := template.New(name)
	tmpl.Funcs(renderFuncs(tmpl))
	tmpl.Funcs(globals(r))

	var err error
	namespace := filepath.Join(filepath.Dir(name), "*.html")
	tmpl, err = tmpl.ParseFS(r.templates, append(r.globalPaths, namespace)...)
	if err != nil {
		return nil, fmt.Errorf("parse templates from filesystem for namespace %q", namespace)
	}

	return tmpl, nil
}
