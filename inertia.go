package inertia

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

type inertiaCtxKeyType string

const inertiaCtxKey inertiaCtxKeyType = "inertia"

type Inertia struct {
	RootTemplate *template.Template

	Version     string
	VersionFunc func() string
	shared      P
	sharedLazy  map[string]func(r *http.Request) interface{}
}

func (i *Inertia) Share(p P) {
	if i.shared == nil {
		i.shared = P{}
	}
	i.shared.merge(p)
}

func (i *Inertia) ShareLazy(prop string, lazy func(r *http.Request) interface{}) {
	i.sharedLazy[prop] = lazy
}

type P map[string]interface{}

// merge merges two maps. On duplicates, if two maps merge recursively, replace with other's key otherwise.
func (p P) merge(other P) {
	for k, v := range other {
		existing, ok := p[k]
		if ok {
			existingP, ok1 := existing.(P)
			vP, ok2 := v.(P)
			if ok1 && ok2 {
				existingP.merge(vP)
			} else {
				p[k] = v
			}
		} else {
			p[k] = v
		}
	}
}

func (i *Inertia) getVersion() string {
	if i.VersionFunc != nil {
		return i.VersionFunc()
	}
	return i.Version
}

type page struct {
	Component string                 `json:"component"`
	Props     map[string]interface{} `json:"props"`
	URL       string                 `json:"url"`
	Version   string                 `json:"version"`
}

func (i *Inertia) render(w http.ResponseWriter, r *http.Request, componentName string, p P) {

	props := P{}
	props.merge(i.shared)

	for prop, lazy := range i.sharedLazy {
		props[prop] = lazy(r)
	}

	props.merge(p)

	if only := strings.Split(r.Header.Get("X-Inertia-Partial-Data"), ","); len(only) != 0 && r.Header.Get("X-Inertia-Partial-Component") == componentName {
		newProps := make(map[string]interface{})
		for _, k := range only {
			if p, ok := props[k]; ok {
				newProps[k] = p
			}
		}
		props = newProps
	}

	// perform lazy evaluation
	for k, v := range props {
		if f, ok := v.(func() interface{}); ok {
			props[k] = f()
		}
	}

	page := page{
		Component: componentName,
		Props:     props,
		URL:       r.RequestURI,
		Version:   i.getVersion(),
	}

	marshalled, err := json.Marshal(page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.Header.Get("X-Inertia") != "" {
		w.Header().Add("Vary", "Accept")
		w.Header().Add("X-Inertia", "true")
		w.Header().Add("Content-type", "application/json")
		w.Write(marshalled)
		return
	}

	err = i.RootTemplate.Execute(w, map[string]interface{}{
		"page": template.HTML(marshalled),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func Render(w http.ResponseWriter, r *http.Request, componentName string, props P) {
	inertia, ok := r.Context().Value(inertiaCtxKey).(*Inertia)
	if !ok {
		panic("[Inertia] No middleware configured.")
	}

	inertia.render(w, r, componentName, props)
}
