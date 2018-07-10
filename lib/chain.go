package lib

import (
	"net/http"
	"io/ioutil"
    "fmt"
    //"strings"
    "errors"
    "bytes"
    //"encoding/json"
    "github.com/Jeffail/gabs"
)


func handleChain(w http.ResponseWriter, req *http.Request, pipes []Chain) error {
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    req.Body = ioutil.NopCloser(bytes.NewReader(body))


    result:= gabs.New()
    fmt.Println("","BeforePipe")
    for _, pipe := range pipes {
        proxyScheme:="http"
        proxyHost:=pipe.Service 
        proxyUrl:= pipe.Endpoint

        httpClient:=http.Client{}
        url := fmt.Sprintf("%s://%s%s", proxyScheme, proxyHost, proxyUrl)
        proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))
        proxyReq.Header = make(http.Header)
        for h, val := range req.Header {
            proxyReq.Header[h] = val
        }

        resp, err := httpClient.Do(proxyReq)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadGateway)
            return errors.New("cannot get data")
        }

        bodyBytes, _ := ioutil.ReadAll(resp.Body)

        if (resp.StatusCode!=200) {
            return errors.New(string(bodyBytes))
        }

        jsonItem, parseErr := gabs.ParseJSON(bodyBytes)
        //jsonItem, parseErr := gabs.ParseJSON([]byte("[{\"foo\":\"bar\"}]"))
        if (parseErr!=nil) {
            fmt.Println("parseErr",parseErr)    
        }
        fmt.Println(jsonItem.String())
        //jsonItem, err := gabs.ParseJSON([]byte("{\"foo\":\"bar\"}"))
        
        mergeErr:= result.Merge(jsonItem)
        if (mergeErr!=nil) {
            fmt.Println("mergeErr ",err)
        }

        defer resp.Body.Close()
    } 
    
    w.Write([]byte(result.String()))
    return nil
}
