package lib

import (
	"encoding/json"
	"net/http"
	//"net/url"
	"io/ioutil"
	"net/http/httputil"
	"strings"
    "errors"
    "fmt"
)






// Find service matches by the url pattern
// returns url, service, error
func backend(c *Config, r *http.Request) (*Service, *string, error) {
	var (
		route string
        query string
	)

    ps := strings.Split(r.URL.Path, "/")
    fmt.Println("ps: ", ps)
    // compare path with each rule
	for rule, service := range c.Rules {
        route="/"
        for index, part:= range ps {
            if (part == "") {
                continue
            }          
            route+= part + "/"

            if strings.Index(route, rule) == 0 {
                query=strings.Join(ps[index+1:],"/")
                return &service, &query, nil
            }
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





func defaultMod (res *http.Response) error {
    bodyBytes, _ := ioutil.ReadAll(res.Body)
    bodyString := string(bodyBytes)
    fmt.Println("default response", bodyString)
    return nil
}



// New creates a new gateway.
func New(c *Config, middlewares map[string] func(http.ResponseWriter, *http.Request) ( *http.Request , error) ) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
        fmt.Println("Some input request got: ",req.Host, req.URL)
		service, query, err := backend(c, req)
		if err!=nil {
			resp, _ := json.Marshal(c.NotFoundResponse)
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-type", "application/json")
			w.Write(resp)
            w.Write([]byte(err.Error()))
			return
		}
        if (middlewares != nil) {
            for _, filterName := range service.Filters {
               if md, ok := middlewares[filterName]; ok {
                   request_out, err:= md(w, req ); 
                   if (err == nil ) {
                        req = request_out
                   } else {
                        // here we must stop everything
                        w.WriteHeader(http.StatusNotFound)
                        w.Header().Set("Content-type", "application/json")
                        w.Write([]byte(err.Error()))
                        return
                   }
               }
            }
        }
        fmt.Printf("one")
        fmt.Printf("",service) 
        fmt.Printf("",service.Aggregate) 
        if (len(service.Aggregate)!=0) {
            fmt.Printf("Hanndle aggregation")
            handleAggregate(w, req, service.Aggregate)
            return;
        }
        fmt.Printf("two")

        if (len(service.Pipes)!=0) {
            pipeline:=Pipeline{Index:0, Pipes:service.Pipes}
            pipeline.BuildProxyPipe(w, c, req)
            return
        }
        fmt.Printf("three")

        if (service.Service!="") {
            (&httputil.ReverseProxy{
                Director: func(r *http.Request) {
                    r.URL.Scheme = c.Scheme//"http"
                    r.URL.Host = service.Service
                    r.URL.Path = "/"+*query 
                    r.Host = service.Service
                },
            }).ServeHTTP(w, req)
        }
	}
}
