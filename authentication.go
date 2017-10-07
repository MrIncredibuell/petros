package petros

import (
	"net/http"

	"github.com/MrIncredibuell/petros/user"

	"github.com/gorilla/securecookie"
)

type AuthenticationSettings struct {
	CookieName   string
	HashKey      []byte
	BlockKey     []byte
	secureCookie *securecookie.SecureCookie
}

func (s *Server) Authenticated(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		if s.GetUser(req) != nil {
			f(w, req)
			return
		}
		http.Redirect(w, req, "/login/", http.StatusFound)
	}
}

func (s *Server) GetUser(req *http.Request) *user.User {
	if u := req.Context().Value(currentUserKey); u != nil {
		return u.(*user.User)
	}

	if cookie, err := req.Cookie(s.Config.AuthenticationSettings.CookieName); err == nil {
		value := make(map[string]string)
		if err = s.Config.AuthenticationSettings.secureCookie.Decode(s.Config.AuthenticationSettings.CookieName, cookie.Value, &value); err == nil {
			return s.UserStore.GetUser("Id", value["id"])
			// req.Context = context.WithValue(req.Context, currentUserKey, u)
		}
	}

	return nil
}

func (s *Server) SetUser(w http.ResponseWriter, req *http.Request, u *user.User) error {
	value := make(map[string]string)
	age := -1
	if u != nil {
		value["id"] = u.Id
		value["username"] = u.Username
		age = 60 * 60 * 24 * 365
	}
	if encoded, err := s.Config.AuthenticationSettings.secureCookie.Encode(s.Config.AuthenticationSettings.CookieName, value); err == nil {
		cookie := &http.Cookie{
			Name:  s.Config.AuthenticationSettings.CookieName,
			Value: encoded,
			Path:  "/",
			// Secure: true,
			HttpOnly: true,
			MaxAge:   age,
		}
		// fmt.Println(cookie)
		http.SetCookie(w, cookie)
	} else {
		return err
	}

	return nil
}
