package main

import (
	"log"
	"errors"
	"plugin"
	"net/http"
	"os"
    . "gateway/lib"
)

func ENV(name string) string {
    result:=""
    if s, ok := os.LookupEnv(name); ok {
        result = s
    } else {
        log.Fatal("Could not get env var " +  name)
    }
    return result
}

type MiddlewarePlugin interface {
    Init() func( http.ResponseWriter, *http.Request ) ( *http.Request, error ) 
}

func LoadPlugin(path string) ( MiddlewarePlugin, error ) {
    mod, err:=plugin.Open(path)
    if err!=nil {
         return nil, err
    }
    
    plug, err:= mod.Lookup("Plugin")
    if (err!=nil) {
        return nil, err
    }
    plugin, ok := plug.(MiddlewarePlugin)
    if (!ok) {
       errors.New("Could not cast to Middleware plugin") 
    }
    return plugin, nil
}

func main() {
	var (
		configPath string
		port       string
	)

	middlewares := make(map[string]func(http.ResponseWriter, *http.Request) ( *http.Request , error))
	if len(os.Args) != 2 {
		if s, ok := os.LookupEnv("GATEWAY_CONFIG_FILE"); ok {
			configPath = s
		} else {
			log.Fatal("Usage: gateway path-to-config.json")
		}
	} else {
		configPath = os.Args[1]
	}

	c, err := LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

    for plugin_name, plugin_path:= range c.Middleware {
        plugin, err:=LoadPlugin(plugin_path)
        if err==nil {
            middlewares[plugin_name]=plugin.Init()
        }        
    }

	if c.Port == "" {
		p, ok := os.LookupEnv("HTTP_PLATFORM_PORT")
		if !ok {
			log.Fatal("Config file should specify port, or the HTTP_PLATFORM_PORT environment variable must be set.")
		}
		port = p
	} else {
		port = c.Port
	}

    http.HandleFunc("/init/", func(w http.ResponseWriter, r *http.Request) {
        r.ParseForm()
    })

	http.HandleFunc("/", New(c, middlewares))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
