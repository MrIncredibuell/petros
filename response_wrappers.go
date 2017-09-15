package petros

import "net/http"

type StatusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *StatusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *StatusRecorder) GetStatus() int {
	return w.status
}

func RecordStatus(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{ResponseWriter: w, status: http.StatusOK}
}

type BodyDropper struct {
	http.ResponseWriter
}

func (w *BodyDropper) Write(body []byte) (int, error) {
	return len(body), nil
}

func DropBody(w http.ResponseWriter) *BodyDropper {
	return &BodyDropper{w}
}
