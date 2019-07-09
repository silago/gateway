package lib

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

/* Gateway Api reverse proxy  that handles tcp socket, http and websocket protocols*/
type Gateway struct {
	config     *Config
	middleware map[string]PluginInterface
	configPath string
}

func (gw *Gateway) checkIsConfigReady() {
	if gw.config == nil {
		panic("could not run gateway without gw.config")
	}
}

/* StartReloadSignal waits for signal from os an rereads config */
func (gw *Gateway) StartReloadOnSignal(sig syscall.Signal) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sig)
	go func() {
		for sig := range sigChan {
			log.Printf("got {%s}  signal. reloading config ", sig)
			if newConfig, err := LoadConfig(gw.configPath); err == nil {
				*gw.config = *newConfig
			} else {
				log.Printf(err.Error())
				os.Exit(1)
			}
		}
	}()
	return
}

//func (gw *Gateway) Start() {
//	gw.checkIsConfigReady()
//	http.HandleFunc("/", gw.GetHandler())
//}

func (gw *Gateway) loadMiddleware() {
	gw.checkIsConfigReady()
	for pluginName, pluginPath := range gw.config.Middleware {
		plugin, err := LoadPlugin(pluginPath)
		if err == nil {
			gw.middleware[pluginName] = plugin.Init()
		} else {
			fmt.Println("cant load plugin", pluginPath, err.Error())
		}
	}
}


func handleTcpProxyConnection(conn net.Conn, targetServer *net.TCPAddr) {
	client, dialError := net.DialTCP("tcp", nil, targetServer);
	if dialError != nil {
		_ = conn.Close()
		log.Println(dialError.Error())
		return
	}
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
			defer wg.Done()
			defer client.Close()
			defer conn.Close()

			for {
				inputBuffer := make([]byte, 256)
				_ = client.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
				n, e := client.Read(inputBuffer);
				if err, ok := e.(net.Error); ok && err.Timeout() {
					continue
				}
				if e==nil {
					if _, e = conn.Write([]byte(string(inputBuffer[:n]))); e !=nil {
						return
					}
				} else {
					return
				}
		}
	}()

	go func() {
			defer wg.Done()
			defer client.Close()
			defer conn.Close()
			for {
				inputBuffer := make([]byte, 256)
				_ = conn.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
				n, e := conn.Read(inputBuffer)
				if err, ok := e.(net.Error); ok && err.Timeout() {
					continue
				}
				if e == nil  {
					if _, e = client.Write([]byte(string(inputBuffer[:n]))); e != nil {
						return
					}
				} else {
					return
				}

		}
	}()
	wg.Wait()
	log.Println("tcp proxy closed")
}

func (gw *Gateway) StartTcpPortForwarding() {
	for host, target := range gw.config.PortForward {
		var addr , resolveError = net.ResolveTCPAddr("tcp",host)
		if resolveError!=nil {
			panic(resolveError.Error())
		}
		listener, listenError := net.ListenTCP("tcp", addr)
		defer listener.Close()
		if listenError != nil {
			log.Println(listenError.Error())
			return
		}

		targetServer, _:=net.ResolveTCPAddr("tcp",target)
		go func() {
			for {
				conn, acceptError := listener.Accept()
				if acceptError != nil {
					log.Println(acceptError.Error())
					_ = conn.Close()
					continue
				}
				go handleTcpProxyConnection(conn,targetServer)
			}
		}()
		}
		select {}
	}


func InitGateway(configPath string) *Gateway {
	var gw = &Gateway{configPath: configPath}
	if config, err := LoadConfig(configPath); err != nil {
		panic(err)
	} else {
		gw.config = config
	}
	return gw
}

func (gw *Gateway) GetHandler() http.HandlerFunc { //(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		service, query, route_error := gw.getTargetService(req)
		if route_error == nil {
			switch proto := service.Protocol; proto {
			case "ws":
				if err := WebsocketProxyHandler(w, req, service, query, gw.middleware); err != nil {
					log.Println("[error][websocket]:", err.Error())
					w.Write([]byte(err.Error()))
				}
			default:
				if err := HttpProxyHandler(w, req, service, query, gw.middleware); err != nil {
					w.Write([]byte(err.Error()))
				}
			}
		} else {
			w.Write([]byte("{\"error\":\"page not found\"}"))
		}
	}
}

//func New(c *Config, middlewares map[string]func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)) http.HandlerFunc {
//	return func(w http.ResponseWriter, req *http.Request) {
//		service, query, route_error := getTargetService(c, req)
//		if route_error == nil {
//			switch proto := service.Protocol; proto {
//			case "ws":
//				if err := wsProxy(w, req, service, query, middlewares); err!=nil {
//					log.Println("[error][websocket]:", err.Error())
//					w.Write([]byte(err.Error()))
//				}
//			default:
//				if err := httpProxy(w, req, service, query, middlewares); err!=nil {
//					log.Println("[error][proxy]: ", err.Error())
//					w.Write([]byte(err.Error()))
//				}
//			}
//		} else {
//			w.Write([]byte("{\"error\":\"page not found\"}"))
//		}
//	}
//}

// Find service matches by the url pattern
// returns url, service, error
func (gw *Gateway) getTargetService(r *http.Request) (*Service, *string, error) {
	var c = gw.config
	var (
		route string
		query string
	)
	ps := strings.Split(r.URL.Path, "/")
	for rule, service := range c.Rules {
		route = "/"
		for index, part := range ps {
			if part == "" {
				continue
			}
			route += part + "/"

			if strings.Index(route, rule) == 0 {
				query = strings.Join(ps[index+1:], "/")
				return &service, &query, nil
			}
		}
	}
	return nil, nil, errors.New("{\"error\":\"route not found\"}")
}

