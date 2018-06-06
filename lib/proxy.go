package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"strings"
    "errors"
)

// returns url, service, error
func backend(c *Config, r *http.Request) (*Service, *string, error) {
	var (
		route string
        query string
	)

	if c.Version != "" {
		ps := strings.SplitN(r.URL.Path, "/", 3)
        route = "/" + ps[1]
        if (len(ps)>2) {
            query=ps[2]
        } else {
            query=""
        }

	} else {
		route = r.URL.Path
	}

    // compare path with each rule
	for rule, service := range c.Rules {
		if strings.Index(route, rule) == 0 {
			return &service, &query, nil
		}
	}
	return nil, nil, errors.New("Route not found")
}


func tryFallback(c *Config, r *http.Request) (string, string, bool){
	if c.Version != "" && c.FallbackRule != "" {
		return c.FallbackRule, r.URL.Path, true
	}
	return "", "", false
}

// New creates a new gateway.
func New(c *Config, middlewares map[string]func(http.ResponseWriter, *http.Request) ( *http.Request , error) ) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
        var (
            outRequest *http.Request
        )
        outRequest = req

		service, query, err := backend(c, req)
		if err!=nil {
			resp, _ := json.Marshal(c.NotFoundResponse)
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-type", "application/json")
			w.Write(resp)
            w.Write([]byte(err.Error()))
			return
		}

        if (middlewares != nil)  {
            for _, filterName := range service.Filters {
               if md, ok := middlewares[filterName]; ok {
                   request_out, err:= md(w, req ); 
                   if (err != nil ) {
                        outRequest = request_out
                   }
               }
            }
        }

		(&httputil.ReverseProxy{
			Director: func(r *http.Request) {
				r.URL.Scheme = "http"
				r.URL.Host = service.Service
				r.URL.Path = "/"+*query 
				r.Host = service.Service
			},
		}).ServeHTTP(w, outRequest)
	}
}
