package petros

import (
	"context"
	"errors"
	"net/http"
)

var ParameterNotFoundError = errors.New("Not Found")

type RequestArgs struct {
	Headers     map[string][]string
	URLParams   map[string]string
	QueryParams map[string][]string
	BodyParams  map[string][]string
}

func ParsedRequest(req *http.Request) *http.Request {
	args, err := Args(req)
	if err != nil {
		// TODO: decide what to do here
		// http.Error(w, "Unable to parse request", http.StatusBadRequest)
		return nil
	}

	req = req.WithContext(context.WithValue(req.Context(), requestArgsKey, args))
	return req
}

func Args(req *http.Request) (*RequestArgs, error) {
	if req == nil {
		return nil, errors.New("Cannot get args for a nil request")
	}

	if value, ok := req.Context().Value(requestArgsKey).(*RequestArgs); value != nil && ok {
		return value, nil
	}

	args := &RequestArgs{
		Headers:     req.Header,
		URLParams:   make(map[string]string),
		QueryParams: req.URL.Query(),
		BodyParams:  make(map[string][]string),
	}

	body := req.Body
	if body != nil {
		if contentType, found := args.Headers["Content-Type"]; found {
			if contentType[0] == "application/x-www-form-urlencoded" {
				req.ParseForm()
				args.BodyParams = req.Form
			}
		}
	}

	return args, nil
}

func (a *RequestArgs) GetString(key string) (string, error) {
	if values, found := a.BodyParams[key]; found && len(values) > 0 {
		return values[0], nil
	}

	if values, found := a.QueryParams[key]; found && len(values) > 0 {
		return values[0], nil
	}

	return "", ParameterNotFoundError
}

type PageInfo struct {
	Title       string
	StyleSheets []string
}

func (server *Server) TemplateArgs(req *http.Request, info *PageInfo) map[string]interface{} {
	m := make(map[string]interface{})

	m["CurrentUser"] = server.GetUser(req)
	m["PageInfo"] = info
	m["RequestArgs"], _ = Args(req)

	return m
}
