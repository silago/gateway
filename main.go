package main

import (
	"errors"
	"fmt"
	"gateway/lib"
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
		//fmt.Println("could not cast to Middleware plugin")
		return nil, errors.New("could not cast to Middleware plugin")
	}
	return plugin, nil
}

func main() {
	var (
		configPath string
	)
	port := ENV("PORT")
	sslPort := ENV("SSL_PORT")
	middleware := make(map[string]func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error))
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

	for pluginName, pluginPath := range config.Middleware {
		plugin, err := LoadPlugin(pluginPath)
		if err == nil {
			middleware[pluginName] = plugin.Init()
		} else {
			fmt.Println("cant load plugin", pluginPath)
			fmt.Println(err)
		}
	}

	//log.Println("..")
	for host, target := range config.PortForward {
		listener, err := net.Listen("tcp", host)
		defer listener.Close()
		if err != nil {
			log.Println(err.Error())
		}

		//log.Println("listening tcp at", host)
		go func() {
			for {
				if conn, err := listener.Accept(); err!=nil {
					log.Println(err.Error())
					continue
				} else if client, err := net.Dial("tcp", target); err!=nil {
					conn.Close()
					log.Println(err.Error())
					continue
				} else {
					go func() {
						for {
							inputBuffer := make([]byte, 256)
							if n, e := client.Read(inputBuffer); e != nil {
								_ = conn.Close()
								return
							} else {
								message := string(inputBuffer[:n])
								_, _ = conn.Write([]byte(message))
							}
						}
					}()

					go func() {
						for {
							inputBuffer := make([]byte, 256)
							if n, e := conn.Read(inputBuffer); e != nil {
								_ = conn.Close()
								return
							} else {
								message := string(inputBuffer[:n])
								_, _ = client.Write([]byte(message))
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

	http.HandleFunc("/", lib.New(config, middleware))
	log.Println("listen http at ", port)
	go func() {
		log.Fatal(http.ListenAndServeTLS(":"+sslPort, "server.crt", "server.key", nil))
	}()
	log.Fatal(http.ListenAndServe(":"+port, nil))
	select {}
}
