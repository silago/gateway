package main

import (
	"log"
	"net/http"
)

const criticalResponseTime float64 = 0.5 // seconds
type timelog struct {
}

func (s timelog) after(response *http.Response) (*http.Response, error) {
	return response, nil
}
//    in: --
//    out:
//    	func:
//			out: r, e
//			in:
//				- req,
//				- func:
//					- in: req
//					- out: res, err

func (s timelog) Init() func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	log.Println("init time check plugin")
	return func(request *http.Request, wrapped func(*http.Request) (*http.Response, error)) (*http.Response, error) {
		//start:= time.Now()
		wrappedResponse, _ := wrapped(request)
		response, err := s.after(wrappedResponse)

		//elapsed := time.Since(start).Seconds()
		//if elapsed > criticalResponseTime {
		//	log.Println("[W] request took too long: ",elapsed,"sec. ", request.Host, request.URL)
		//}
		return response, err
	}
}

var Plugin timelog
