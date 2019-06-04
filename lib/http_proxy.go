package lib

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"
)

const criticalResponseTime float64 = 0.4

func GetDefaultHandler(service *Service, query *string) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (response *http.Response, e error) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		url := fmt.Sprintf("%s://%s/%s", service.Scheme, service.Service, *query)
		proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		proxyReq.Header = make(http.Header)
		proxyReq.URL.RawQuery = req.URL.RawQuery

		for h, val := range req.Header {
			proxyReq.Header[h] = val
		}

		//httpClient := http.Client{}

		//resp, err := httpClient.Do(proxyReq)

		var start, connect, dns, tlsHandshake time.Time
		var logs map[string]string = make(map[string]string)

		trace := &httptrace.ClientTrace{
			DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
			DNSDone: func(ddi httptrace.DNSDoneInfo) {
				logs["dnsdone"] = fmt.Sprintf("DNS Done: %v", time.Since(dns).Seconds())
			},

			TLSHandshakeStart: func() { tlsHandshake = time.Now() },
			TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
				logs["handshake"] = fmt.Sprintf("TLS Handshake: %v", time.Since(tlsHandshake).Seconds())
			},

			ConnectStart: func(network, addr string) { connect = time.Now() },
			ConnectDone: func(network, addr string, err error) {
				logs["connectTime"] = fmt.Sprintf("Connect time: %v", time.Since(connect).Seconds())
			},
			GotFirstResponseByte: func() {
				logs["starttofirst"] = fmt.Sprintf("Time from start to first byte: %v", time.Since(start).Seconds())
			},
		}
		proxyReq = proxyReq.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		proxyReq.Close = true

		start = time.Now()
		if resp, err := http.DefaultTransport.RoundTrip(proxyReq); err != nil {
			return nil, err
		} else {
			var total = time.Since(start).Seconds()
			if total > criticalResponseTime {
				log.Println("[W] too long response: ", total, req.Host, req.URL)
				for k, v := range logs {
					log.Println("[W]", req.Host, k, ":", v)
				}
				log.Println("")
			}
			return resp, nil
		}
	}
}

func HttpProxyHandler(w http.ResponseWriter, req *http.Request, service *Service, query *string,
	middlewares map[string]PluginInterface) error {
	handler := GetDefaultHandler(service, query)
	if middlewares != nil {
		for _, filterName := range service.Plugins {
			chainElement := handler
			if plugin, ok := middlewares[filterName]; ok {
				//lib.				pluginMethod = plugin.(func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error))
				handler = func(req *http.Request) (*http.Response, error) {
					if _resp, err := plugin(req, chainElement); err != nil {
						return _resp, err
					} else if _resp != nil {
						return _resp, err
					} else {
						return nil, errors.New(fmt.Sprint("got empty response from: ", req.URL))
					}
				}
			} else {
				log.Println(" cant load plugin ")
			}
		}
	}

	if req == nil {
		return errors.New("nil request")
	}

	resp, err := handler(req)
	if err != nil {
		return err
	}

	if resp.Header != nil {
		for responseHeader, responseHeaderValues := range resp.Header {
			for _, responseHeaderValue := range responseHeaderValues {
				w.Header().Add(responseHeader, responseHeaderValue)
			}
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return err
	} else {
		defer resp.Body.Close()
		_, _ = w.Write(body)
	}
	return nil
}
