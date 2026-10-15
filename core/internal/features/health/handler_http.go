package health

import "net/http"

type HTTPHandler struct{}

func NewHTTP() *HTTPHandler {
	return &HTTPHandler{}
}

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", h.healthCheck)
}

func (h *HTTPHandler) healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
