package lib

import (
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type msg struct {
	Message []byte
	Type    int
}

/*WebsocketProxyHandler  forwards requests and responses  from  client to target service by http */
func WebsocketProxyHandler(w http.ResponseWriter, req *http.Request, service *Service, query *string, middlewares map[string]PluginInterface) error {
	conn, err := upgrader.Upgrade(w, req, nil) // error ignored for sake of simplicity

	if err!=nil {
		log.Println(err)
		return err
	}
		go func(connection *websocket.Conn, req *http.Request, service *Service, path string) {
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

			serviceConn, _, err := dialer.Dial(u.String(), requestHeader)
			if err != nil {
				log.Println(err)
				return
			}

			defer func() {
				_ = connection.Close()
				_ = serviceConn.Close()
			}()

			errorChan := make(chan error)
			clientChan := make(chan msg)
			serviceChan := make(chan msg)

			go startHostListener(connection, serviceChan, errorChan)
			go startServiceListener(serviceConn, clientChan, errorChan)
			startProxy(connection, serviceConn, serviceChan, clientChan,errorChan )
		}(conn, req, service, *query)
		return nil
}


func startHostListener(conn *websocket.Conn, serviceChan chan msg, errorChan chan error) {
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
}

func startServiceListener(serviceConn *websocket.Conn, clientChan chan  msg, errorChan chan error) {
	for {
		select {
		default:
			messageType, message, err := serviceConn.ReadMessage()
			if err == nil {
				clientChan <- msg{Message: message, Type: messageType}
			} else {
				errorChan <- err
			}
		}
	}
}


func startProxy(connection *websocket.Conn, serviceConn *websocket.Conn, serviceChan chan msg, clientChan chan msg, errorChan chan error) {
	for {
		select {
		case message := <-clientChan:
			writeError := connection.WriteMessage(message.Type, message.Message)
			if writeError != nil {
				errorChan <- writeError
			}
		case message := <-serviceChan:
			writeError := serviceConn.WriteMessage(message.Type, message.Message)
			if writeError != nil {
				errorChan <- writeError
			}
		case err := <-errorChan:
			log.Println(err.Error())
			return
		}
	}
}
