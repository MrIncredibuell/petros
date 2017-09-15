package petros

import (
	"net/http"
	"strings"
)

func (server *Server) AllowMethods(methods map[string]func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	cannonicalMethods := make(map[string]func(http.ResponseWriter, *http.Request))
	for key, value := range methods {
		cannonicalMethods[strings.TrimSpace(strings.ToUpper(key))] = value
	}

	allowed := []string{}
	for key := range cannonicalMethods {
		allowed = append(allowed, key)
	}

	_, found := cannonicalMethods["HEAD"]
	if !found {
		allowed = append(allowed, "HEAD")
		cannonicalMethods["HEAD"] = func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			f, found := cannonicalMethods["GET"]
			if found {
				dropper := DropBody(w)
				f(dropper, req)
			} else {
				w.WriteHeader(200)
			}
			return
		}
	}

	_, found = cannonicalMethods["OPTIONS"]
	if !found {
		allowed = append(allowed, "OPTIONS")
		cannonicalMethods["OPTIONS"] = func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			w.WriteHeader(200)
			return
		}
	}

	return func(w http.ResponseWriter, req *http.Request) {
		f, found := cannonicalMethods[req.Method]
		if found {
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			f(w, req)
			return
		}

		w.Header().Set("Allow", strings.Join(allowed, ", "))
		w.WriteHeader(405)
	}
}
