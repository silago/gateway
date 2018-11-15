package main

import (
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type test struct {
}

func (s test) Init() func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	return  func(request *http.Request,   wrapped func(*http.Request) (*http.Response, error)) (*http.Response, error) {
		return wrapped(request)
	}
}

var Plugin test