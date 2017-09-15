package petros

import (
	// "net/http"
	"errors"
	"html/template"
	"io"
	"path"
	"sync"
)

type TemplateMux struct {
	sync.Mutex
	root               string
	defaultFunctionMap template.FuncMap
	m                  map[string]*template.Template
}

func (t *TemplateMux) AddTemplate(name string, filenames ...string) error {
	for i, fname := range filenames {
		if !path.IsAbs(fname) {
			filenames[i] = path.Join(t.root, fname)
		}
	}

	temp, err := template.New("").Funcs(t.defaultFunctionMap).ParseFiles(filenames...)
	if err != nil {
		return err
	}

	t.Lock()
	defer t.Unlock()

	t.m[name] = temp

	return nil
}

func (t *TemplateMux) Render(name string, w io.Writer, args interface{}) error {
	t.Lock()
	defer t.Unlock()

	temp, found := t.m[name]
	if !found {
		return errors.New("Template not found: " + name)
	}

	return temp.ExecuteTemplate(w, "base", args)

}

func (t *TemplateMux) AddDefaultFunctions(fs template.FuncMap) error {
	for key, value := range fs {
		t.defaultFunctionMap[key] = value
	}
	return nil
}

func NewTemplateMux(root string, defaultFunctions template.FuncMap) *TemplateMux {
	if defaultFunctions == nil {
		defaultFunctions = make(template.FuncMap)
	}
	return &TemplateMux{
		m:                  make(map[string]*template.Template),
		root:               root,
		defaultFunctionMap: defaultFunctions,
	}
}
