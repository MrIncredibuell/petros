package petros

import "net/http"

func (s *Server) HandlePanic(f func(http.ResponseWriter, *http.Request), errorFunc func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				s.ErrLog.Println("Recovered in f", r)
				errorFunc(w, req)
			}
		}()

		f(w, req)
	}
}
