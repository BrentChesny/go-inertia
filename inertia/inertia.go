package inertia

import (
	"encoding/json"
	"html/template"
	"net/http"
)

type inertiaCtxKeyType string

const inertiaCtxKey inertiaCtxKeyType = "inertia"

type Inertia struct {
	RootTemplate *template.Template

	Version     string
	VersionFunc func() string
}

func (i *Inertia) getVersion() string {
	if i.VersionFunc != nil {
		return i.VersionFunc()
	}
	return i.Version
}

type page struct {
	Component string      `json:"component"`
	Props     interface{} `json:"props"`
	URL       string      `json:"url"`
	Version   string      `json:"version"`
}

func (i *Inertia) render(w http.ResponseWriter, r *http.Request, componentName string, props interface{}) {
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

func Render(w http.ResponseWriter, r *http.Request, componentName string, props interface{}) {
	inertia, ok := r.Context().Value(inertiaCtxKey).(*Inertia)
	if !ok {
		panic("[Inertia] No middleware configured.")
	}

	inertia.render(w, r, componentName, props)
}
