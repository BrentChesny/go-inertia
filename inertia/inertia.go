package inertia

import (
	"encoding/json"
	"html/template"
	"net/http"
)

var (
	rootTmpl *template.Template

	version     string
	versionFunc func() string
)

type page struct {
	Component string      `json:"component"`
	Props     interface{} `json:"props"`
	URL       string      `json:"url"`
	Version   string      `json:"version"`
}

func Init(rootTemplate *template.Template) {
	rootTmpl = rootTemplate
}

func Version(v string) {
	version = v
}

func VersionFunc(f func() string) {
	versionFunc = f
}

func GetVersion() string {
	if versionFunc != nil {
		return versionFunc()
	}
	return version
}

func Render(w http.ResponseWriter, r *http.Request, componentName string, props interface{}) {
	page := page{
		Component: componentName,
		Props:     props,
		URL:       r.RequestURI,
		Version:   GetVersion(),
	}

	marshalled, err := json.Marshal(page)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	if r.Header.Get("X-Inertia") != "" {
		w.Header().Add("Vary", "Accept")
		w.Header().Add("X-Inertia", "true")
		w.Header().Add("Content-type", "application/json")
		w.Write(marshalled)
		return
	}

	err = rootTmpl.Execute(w, map[string]interface{}{
		"page": template.HTML(marshalled),
	})
	if err != nil {
		w.WriteHeader(500)
		return
	}
}

func Middleware() {
	// TODO
}
