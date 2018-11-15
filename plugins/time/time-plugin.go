package main

import (
	"fmt"
	"net/http"
	"time"
)

const TIME_HEADER = "X-Server-Time"

type servertime struct {
}

func (s servertime) after(response *http.Response) (*http.Response, error) {
	response.Header.Set(TIME_HEADER, fmt.Sprint("", int(time.Now().Unix())))
	//response.Header.Add(TIME_HEADER, fmt.Sprintf("",time.Now().Unix()))
	return response, nil
}

func (s servertime) Init() func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	return func(request *http.Request, wrapped func(*http.Request) (*http.Response, error)) (*http.Response, error) {
		wrapped_response, _ := wrapped(request)
		response, err := s.after(wrapped_response)
		return response, err
	}
}

var Plugin servertime
