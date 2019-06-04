package lib

import (
	"errors"
	"net/http"
	"plugin"
)

type PluginInterface func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)

type MiddlewarePlugin interface {
	Init() PluginInterface //func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)
	//Init() func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)
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
