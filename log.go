package petros

import (
	"net/http"
	"time"
)

func (s *Server) LogRequest(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		wrapped := RecordStatus(w)
		f(wrapped, req)
		status := wrapped.GetStatus()
		user := s.GetUser(req)
		username := "Anonymous"
		if user != nil {
			username = "@" + user.Username
		}
		s.Log.Println(status, req.Method, req.URL, username, time.Now().Sub(start))
	}
}
