package petros

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/MrIncredibuell/petros/user"
	"github.com/gorilla/securecookie"
)

type Config struct {
	AuthenticationSettings AuthenticationSettings
	StaticRoot             string
	UserStoreFilename      string
	TemplateRoot           string
}

type Server struct {
	httpServer http.Server
	Log        *log.Logger
	ErrLog     *log.Logger
	*TemplateMux
	Config    *Config
	UserStore user.UserStore
}

func (s *Server) ListenAndServe() {
	s.httpServer.ListenAndServe()
}

func ParseServerConfig(filename string) *Server {
	var err error
	s := &Server{
		Log:    log.New(os.Stdout, "", 0),
		ErrLog: log.New(os.Stderr, "", 0),
	}

	if file, err := ioutil.ReadFile(filename); err == nil {
		err = json.Unmarshal(file, &s.Config)
		if err != nil {
			panic(err)
		}
	}

	s.httpServer = http.Server{
		Addr: ":5000",
		// Handler: http.HandlerFunc(NewServeMux(server).ServeHTTP),
	}

	// err := InitTemplates(server)
	// if err != nil {
	// 	panic(err)
	// }

	s.TemplateMux = NewTemplateMux(s.Config.TemplateRoot, nil)

	s.UserStore, err = user.NewFileStore(s.Config.UserStoreFilename)
	if err != nil {
		panic(err)
	}

	s.Config.AuthenticationSettings.secureCookie = securecookie.New(
		s.Config.AuthenticationSettings.HashKey,
		s.Config.AuthenticationSettings.BlockKey)

	return s
}

func (s *Server) SetHandler(h http.Handler) {
	s.httpServer.Handler = h
}
