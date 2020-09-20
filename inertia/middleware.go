package inertia

import (
	"context"
	"net/http"
)

type MiddlewareFunc func(http.Handler) http.Handler

func Middleware(inertia *Inertia) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), inertiaCtxKey, inertia)
			req := r.WithContext(ctx)

			if r.Header.Get("X-Inertia") == "" {
				next.ServeHTTP(w, req)
				return
			}

			// In case the assets version in the X-Inertia-Version header does not match the current version
			// of assets we have on the server, return a 409 response which will cause Inertia to make a new
			// hard visit.
			if r.Method == "GET" && r.Header.Get("X-Inertia-Version") != inertia.getVersion() {
				w.Header().Add("X-Inertia-Location", r.URL.String())
				w.WriteHeader(http.StatusConflict)
				return
			}

			rw := &responseWriter{w, req}
			next.ServeHTTP(rw, req)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	req *http.Request
}

func (rw *responseWriter) WriteHeader(status int) {
	if status == http.StatusFound {
		switch rw.req.Method {
		case "PUT", "PATCH", "DELETE":
			rw.WriteHeader(http.StatusSeeOther)
			return
		}
	}
	rw.WriteHeader(status)
}
