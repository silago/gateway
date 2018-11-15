package main

import (
	"errors"
	"fmt"
	lib "gateway/lib"
	"log"
	"net/http"
	"os"
	"plugin"
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

	c, err := lib.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	for plugin_name, plugin_path := range c.Middleware {
		plugin, err := LoadPlugin(plugin_path)
		if err == nil {
			middlewares[plugin_name] = plugin.Init()
		} else {
			fmt.Println("cant load plugin", plugin_path)
			fmt.Println(err)
		}
	}

	http.HandleFunc("/", lib.New(c, middlewares))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
