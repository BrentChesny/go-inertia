package inertia

import (
	"encoding/json"
	"fmt"
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
}

type P map[string]interface{}

type L func() interface{}

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

func (i *Inertia) render(w http.ResponseWriter, r *http.Request, componentName string, props map[string]interface{}) {

	// TODO:  merge shared props

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
		fmt.Println(err)
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
