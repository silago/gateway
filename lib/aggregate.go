package lib

import (
	"net/http"
	"io/ioutil"
    "fmt"
    //"strings"
    "errors"
    "bytes"
    "encoding/json"
    "github.com/Jeffail/gabs"
)


func handleAggregate(w http.ResponseWriter, req *http.Request, pipes map[string]AggregatePipe) error {
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    req.Body = ioutil.NopCloser(bytes.NewReader(body))

    results:= make(map[string]*gabs.Container)
    for index, pipe := range pipes {
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
        var obj *json.RawMessage
        bodyBytes, _ := ioutil.ReadAll(resp.Body)
        str := string(bodyBytes)
        json.Unmarshal([]byte(str), &obj)
        currentObject:=gabs.New()
        currentObject.SetP(obj,index)
        results[index]=currentObject
        defer resp.Body.Close()
    } 
    
    
    result:= gabs.New()
    for _, container:= range results {
        result.Merge(container)
    }
    w.Write([]byte(result.String()))
    return nil
}
