package main

import (
	"errors"
	"fmt"
	lib "gateway/lib"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"plugin"
	"syscall"
)

func ENV(name string) string {
	result := ""
	if s, ok := os.LookupEnv(name); ok {
		result = s
	} else {
		log.Fatal("Could not get env var " + name)
	}
	return result
}

type MiddlewarePlugin interface {
	Init() func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)
}

func LoadPlugin(path string) (MiddlewarePlugin, error) {
	mod, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	plug, err := mod.Lookup("Plugin")
	if err != nil {
		return nil, err
	}
	plugin, ok := plug.(MiddlewarePlugin)
	if !ok {
		fmt.Println("Could not cast to Middleware plugin")
		errors.New("Could not cast to Middleware plugin")
	}
	return plugin, nil
}

func main() {
	var (
		configPath string
	)
	port := ENV("PORT")
	ssl_port := ENV("SSL_PORT")
	middlewares := make(map[string]func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error))
	//      func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error))
	if len(os.Args) != 2 {
		if s, ok := os.LookupEnv("GATEWAY_CONFIG_FILE"); ok {
			configPath = s
		} else {
			log.Fatal("Usage: gateway path-to-config.json")
		}
	} else {
		configPath = os.Args[1]
	}

	config, err := lib.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	sigChan :=make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	go func(c *lib.Config) {
		for sig:=range sigChan {
			log.Printf("got {%s}  signal. reloading config ", sig)
			if newConfig, err := lib.LoadConfig(configPath); err == nil {
				*config=*newConfig
			} else {
				log.Printf(err.Error())
				os.Exit(1)
			}
		}
	}(config)

	for plugin_name, plugin_path := range config.Middleware {
		plugin, err := LoadPlugin(plugin_path)
		if err == nil {
			middlewares[plugin_name] = plugin.Init()
		} else {
			fmt.Println("cant load plugin", plugin_path)
			fmt.Println(err)
		}
	}

	for host, target := range config.PortForward {
		listener, err := net.Listen("tcp", host)
		if err != nil {
			log.Println(err.Error())
		}

		go func() {
			for {
				conn, err := listener.Accept()
				defer conn.Close()
				if err != nil {
					log.Println(err)
					continue
				} else {
					log.Println("accepted")
					client, _ := net.Dial("tcp", target)
					defer client.Close()
					go func() {
						for {
							inputBuffer := make([]byte, 256)
							if n, e := client.Read(inputBuffer); e != nil {
								if e == io.EOF {
									return
								}
							} else {
								message := string(inputBuffer[:n])
								conn.Write([]byte(message))
							}
						}
					}()

					go func() {
						for {
							inputBuffer := make([]byte, 256)
							if n, e := conn.Read(inputBuffer); e != nil {
								if e == io.EOF {
									return
								}
							} else {
								message := string(inputBuffer[:n])
								client.Write([]byte(message))
							}
						}
					}()
				}
			}
		}()
	}

	http.HandleFunc("/google13dd0d8dae2fd927.html", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("google-site-verification: google13dd0d8dae2fd927.html"))
	})

	http.HandleFunc("/", lib.New(config, middlewares))
	log.Println("listen http at ", port)
	go func() {
		log.Fatal(http.ListenAndServeTLS(":"+ssl_port, "server.crt", "server.key", nil))
	}()
	log.Fatal(http.ListenAndServe(":"+port, nil))
	//	log.Println("listen ssl at ", ssl_port)
	for {
	}
}
