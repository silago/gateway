package lib

/*
type Pipeline struct {
	Index int
	Pipes []Pipe
}

func (p *Pipeline) Reset() {
	p.Index = 0
}

func (p *Pipeline) BuildProxyPipe(writer http.ResponseWriter, c *Config, req *http.Request) {
	if len(p.Pipes) == 0 {
		return
	}

	currentPipe := p.Pipes[p.Index]
	(&httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = c.Scheme //"http"
			r.URL.Host = req.Host
			r.URL.Path = "/" + currentPipe.Service + "/" + currentPipe.Endpoint //"/"+path//"/foo"
			r.Host = req.Host
		},
		ModifyResponse: p.BuildNextProxyPipe(writer, c, req),
	}).ServeHTTP(writer, req)
}

func (p *Pipeline) BuildNextProxyPipe(writer http.ResponseWriter, c *Config, req *http.Request) func(*http.Response) error {
	p.Index++
	if p.Index >= len(p.Pipes) {
		return nil
	}
	currentPipe := p.Pipes[p.Index]
	previousPipe := p.Pipes[p.Index-1]
	return func(res *http.Response) error {
		(&httputil.ReverseProxy{
			Director: func(r *http.Request) {
				form, _ := url.ParseQuery(req.URL.RawQuery)
				urlvalues := generateResponseMap(previousPipe, r, res)
				for x := range urlvalues {
					form.Add(x, urlvalues[x])
				}

				//form.Add("boo", "far")
				r.URL.RawQuery = form.Encode()

				r.URL.Scheme = c.Scheme //"http"
				r.URL.Host = req.Host
				r.URL.Path = "/" + currentPipe.Service + "/" + currentPipe.Endpoint //"/"+path//"/foo"
				r.Host = req.Host
			},
			ModifyResponse: p.BuildNextProxyPipe(writer, c, req),
		}).ServeHTTP(writer, req)
		return nil
	}
}

func generateResponseMap(pipe Pipe, req *http.Request, res *http.Response) map[string]string {
	result := make(map[string]string)
	var obj map[string]*json.RawMessage
	bodyBytes, _ := ioutil.ReadAll(res.Body)
	str := string(bodyBytes)
	json.Unmarshal([]byte(str), &obj)

	for name, value := range pipe.Map {
		if val, ok := obj[name]; ok {
			bytes, _ := json.Marshal(val)
			result[value] = string(bytes)
		}
	}
	return result
}
*/
