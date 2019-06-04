package lib

import (
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type Protocols struct {
}

func SocketProxyHandler(w http.ResponseWriter, req *http.Request, service *Service, query *string, middlewares map[string]func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)) error {
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
