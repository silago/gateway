package lib

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func wsProxy(w http.ResponseWriter, req *http.Request, service *Service, query *string, middlewares map[string]func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)) error {
	type msg struct {
		Message []byte
		Type    int
	}

	conn, err := upgrader.Upgrade(w, req, nil) // error ignored for sake of simplicity

	if err == nil {
		go func(conn *websocket.Conn, req *http.Request, service *Service, path string) {
			//var mutex = &sync.Mutex{}
			u := url.URL{Scheme: "ws", Host: service.Service, Path: path}

			dialer := websocket.Dialer{}
			requestHeader := http.Header{}
			if origin := req.Header.Get("Origin"); origin != "" {
				requestHeader.Add("Origin", origin)
			}
			for _, prot := range req.Header[http.CanonicalHeaderKey("Sec-WebSocket-Protocol")] {
				requestHeader.Add("Sec-WebSocket-Protocol", prot)
			}
			for _, cookie := range req.Header[http.CanonicalHeaderKey("Cookie")] {
				requestHeader.Add("Cookie", cookie)
			}

			if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
				if prior, ok := req.Header["X-Forwarded-For"]; ok {
					clientIP = strings.Join(prior, ", ") + ", " + clientIP
				}
				requestHeader.Set("X-Forwarded-For", clientIP)
			}

			requestHeader.Set("X-Forwarded-Proto", "http")
			if req.TLS != nil {
				requestHeader.Set("X-Forwarded-Proto", "https")
			}

			service_conn, _, err := dialer.Dial(u.String(), requestHeader)
			if err != nil {
				return
			}

			defer conn.Close()
			defer service_conn.Close()

			error_chan := make(chan error)
			client_chan := make(chan msg)
			service_chan := make(chan msg)

			go func() {
				for {
					select {
					default:
						messageType, message, err := conn.ReadMessage()
						if err == nil {
							service_chan <- msg{Message: message, Type: messageType}
						} else {
							error_chan <- err
						}
					}
				}
			}()

			go func() {
				for {
					select {
					default:
						messageType, message, err := service_conn.ReadMessage()
						if err == nil {
							client_chan <- msg{Message: message, Type: messageType}
						} else {
							error_chan <- err
						}
					}
				}
			}()

			for {
				select {
				case message := <-client_chan:
					write_error := conn.WriteMessage(message.Type, message.Message)
					if write_error != nil {
						error_chan <- write_error
					}
				case message := <-service_chan:
					write_error := service_conn.WriteMessage(message.Type, message.Message)
					if write_error != nil {
						error_chan <- write_error
					}
				case err := <-error_chan:
					log.Println("ws error: ", err)
					return
				}
			}
		}(conn, req, service, *query)
		return nil
	} else {
		return err
	}
}

func httpProxy(w http.ResponseWriter, req *http.Request, service *Service, query *string, middlewares map[string]func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)) error {
	handler := func(req *http.Request) (*http.Response, error) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		url := fmt.Sprintf("%s://%s/%s", service.Scheme, service.Service, *query)
		log.Println(url)
		proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))
		proxyReq.Header = make(http.Header)

		for h, val := range req.Header {
			proxyReq.Header[h] = val
		}

		httpClient := http.Client{}
		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}

	if middlewares != nil {
		for _, filterName := range service.Plugins {
			log.Println(filterName)
			chainElement := handler
			if plugin, ok := middlewares[filterName]; ok {
				handler = func(req *http.Request) (*http.Response, error) {
					_resp, err := plugin(req, chainElement)
					for h, val := range _resp.Header {
						_resp.Header[h] = val
					}

					return _resp, err
				}
			} else {
				log.Println(" cant load plugin ")

			}
		}
	} else {
		log.Println(" no middlewares detected")
	}

	resp, _ := handler(req)

	for response_header, response_header_values := range resp.Header {
		for _, response_header_value := range response_header_values {
			w.Header().Add(response_header, response_header_value)
		}
	}
	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
	return nil

}
