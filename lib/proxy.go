package lib

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


//func tryFallback(c *Config, r *http.Request) (string, string, bool) {
//	if c.Version != "" && c.FallbackRule != "" {
//		return c.FallbackRule, r.URL.Path, true
//	}
//	return "", "", false
//}

//func New(c *Config, middlewares map[string]func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)) http.HandlerFunc {
//	return func(w http.ResponseWriter, req *http.Request) {
//		service, query, routeError := backend(c, req)
//		if routeError == nil {
//			switch proto := service.Protocol; proto {
//			case "ws":
//				if err := websocketProxyHandler(w, req, service, query, middlewares); err!=nil {
//                    log.Println("[error][websocket]:", err.Error())
//					_, _ = w.Write([]byte(err.Error()))
//				}
//			default:
//				if err := httpProxy(w, req, service, query, middlewares); err!=nil {
//					log.Println("[error][proxy]: ", err.Error(), service.Service, query)
//					_, _ = w.Write([]byte(err.Error()))
//				}
//			}
//		} else {
//			log.Println("route error", req.Host, req.URL, routeError.Error())
//			_, _ = w.Write([]byte("{\"error\":\"page not found\"}"))
//		}
//	}
//}
