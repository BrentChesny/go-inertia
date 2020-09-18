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

func (i *Inertia) GetVersion() string {
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

func Render(w http.ResponseWriter, r *http.Request, componentName string, props interface{}) {
	inertia, ok := r.Context().Value(inertiaCtxKey).(*Inertia)
	if !ok {
		panic("[Inertia] No middleware configured.")
	}

	page := page{
		Component: componentName,
		Props:     props,
		URL:       r.RequestURI,
		Version:   inertia.GetVersion(),
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

	err = inertia.RootTemplate.Execute(w, map[string]interface{}{
		"page": template.HTML(marshalled),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
