package lib

import (
	"net/http"
	"io/ioutil"
    "fmt"
    "errors"
    "bytes"
    "encoding/json"
)


//func makeResponse() {
//    response := &http.Response{}
//    response.Body = ioutil.NopCloser(bytes.NewBuffer())
//}

func handleAggregate(w http.ResponseWriter, req *http.Request, pipes map[string]AggregatePipe) ( map[string]*json.RawMessage, error ) {
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        //return
    }

    // you can reassign the body if you need to parse it as multipart
    req.Body = ioutil.NopCloser(bytes.NewReader(body))
    
    result:=make(map[string]*json.RawMessage)
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
            return result, errors.New("cannot get data")
        }

        var obj *json.RawMessage
        bodyBytes, _ := ioutil.ReadAll(resp.Body)
        str := string(bodyBytes)
        json.Unmarshal([]byte(str), &obj)
        result[index]=obj
        defer resp.Body.Close()
    } 
    r , _:= json.Marshal(result)
    w.Write(r)

    return result, nil
}
