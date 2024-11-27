package render

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig/v3"
)

func renderFuncs(t *template.Template) template.FuncMap {
	m := sprig.TxtFuncMap()
	m["yield"] = yieldFunc(t)

	return m
}

func globals(r *Renderer) template.FuncMap {
	return template.FuncMap{
		"global": func(s string) any {
			return r.globals[s]
		},
	}
}

func yieldFunc(tmpl *template.Template) func(name string, data interface{}) (string, error) {
	return func(name string, data interface{}) (string, error) {
		if t := tmpl.Lookup(name); t == nil {
			return "", nil
		}
		buf := bytes.NewBuffer([]byte{})
		err := tmpl.ExecuteTemplate(buf, name, data)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	}
}

func global(globals map[string]interface{}) func(key string) (interface{}, error) {
	return func(key string) (interface{}, error) {
		value, ok := globals[key]
		if !ok {
			return nil, fmt.Errorf("global value for %q not found", key)
		}
		return value, nil
	}
}
