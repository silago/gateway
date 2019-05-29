package lib

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"
)

func socketProxy(w http.ResponseWriter, req *http.Request, service *Service, query *string, middlewares map[string]func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)) error {
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
				log.Println(err)
				return
			}

			defer func() {
				conn.Close()
				service_conn.Close()
			}()
			errorChan := make(chan error)
			clientChan := make(chan msg)
			serviceChan := make(chan msg)

			go func() {
				for {
					select {
					default:
						messageType, message, err := conn.ReadMessage()
						if err == nil {
							serviceChan <- msg{Message: message, Type: messageType}
						} else {
							errorChan <- err
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
							clientChan <- msg{Message: message, Type: messageType}
						} else {
							errorChan <- err
						}
					}
				}
			}()

			for {
				select {
				case message := <-clientChan:
					write_error := conn.WriteMessage(message.Type, message.Message)
					if write_error != nil {
						errorChan <- write_error
					}
				case message := <-serviceChan:
					write_error := service_conn.WriteMessage(message.Type, message.Message)
					if write_error != nil {
						errorChan <- write_error
					}
				case err := <-errorChan:
					log.Println("ws error: ", err)
					return
				}
			}
		}(conn, req, service, *query)
		return nil
	} else {
		log.Println(err)
		return err
	}
}


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
				log.Println(err)
				return
			}

			defer func() {
				_ = conn.Close()
				_ = service_conn.Close()
			}()

			errorChan := make(chan error)
			clientChan := make(chan msg)
			serviceChan := make(chan msg)

			go func() {
				for {
					select {
					default:
						messageType, message, err := conn.ReadMessage()
						if err == nil {
							serviceChan <- msg{Message: message, Type: messageType}
						} else {
							errorChan <- err
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
							clientChan <- msg{Message: message, Type: messageType}
						} else {
							errorChan <- err
						}
					}
				}
			}()

			for {
				select {
				case message := <-clientChan:
					writeError := conn.WriteMessage(message.Type, message.Message)
					if writeError != nil {
						errorChan <- writeError
					}
				case message := <-serviceChan:
					writeError := service_conn.WriteMessage(message.Type, message.Message)
					if writeError != nil {
						errorChan <- writeError
					}
				case err := <-errorChan:
					log.Println("ws error: ", err)
					return
				}
			}
		}(conn, req, service, *query)
		return nil
	} else {
		log.Println(err)
		return err
	}
}

const criticalResponseTime float64 = 0.4
func getDefaultHandler   (service *Service, query *string) func (req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (response *http.Response, e error) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		url := fmt.Sprintf("%s://%s/%s", service.Scheme, service.Service, *query)
		proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))
		if err!=nil {
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
				logs["handshake"]=fmt.Sprintf("TLS Handshake: %v", time.Since(tlsHandshake).Seconds())
			},

			ConnectStart: func(network, addr string) { connect = time.Now() },
			ConnectDone: func(network, addr string, err error) {
				logs["connectTime"]=fmt.Sprintf("Connect time: %v", time.Since(connect).Seconds())
			},
			GotFirstResponseByte: func() {
				logs["starttofirst"]=fmt.Sprintf("Time from start to first byte: %v", time.Since(start).Seconds())
			},
		}
		proxyReq = proxyReq.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		proxyReq.Close = true

		start = time.Now()
		if resp, err := http.DefaultTransport.RoundTrip(proxyReq); err != nil {
			return nil, err
		} else {
			var total = time.Since(start).Seconds()
			if total>criticalResponseTime {
				log.Println("[W] too long response: ",total, req.Host, req.URL)
				for k, v:=range logs {
					log.Println("[W]", req.Host, k,":",v)
				}
				log.Println("")
			}
			return resp, nil
		}
	}
}


func httpProxy(w http.ResponseWriter, req *http.Request, service *Service, query *string, middlewares map[string]func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)) error {
	handler := getDefaultHandler(service, query)

	if middlewares != nil {
		for _, filterName := range service.Plugins {
			chainElement := handler
			if plugin, ok := middlewares[filterName]; ok {
				handler = func(req *http.Request) (*http.Response, error) {
					if _resp, err := plugin(req, chainElement); err!=nil {
						return _resp, err
					} else if _resp!=nil {
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
		log.Println( err.Error())
        return err
	} else {
		defer  resp.Body.Close()
		_, _ = w.Write(body)
	}
	return nil
}
