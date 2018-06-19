package main

import (
    "encoding/json"
    "fmt"    
    "errors"    
    "sort"
    "strings"
    "crypto/md5"
    "encoding/hex"
	"net/http"
    "io/ioutil"
)




type signer struct {
}

func (s signer) prepareArray(controller string, action string, input_params map[string]interface{}) ( map[string]interface{}, error ) {
    result_params:=input_params["params"].(map[string]interface{})
    added_params:=make(map[string]bool)
    added_params["suid"]          =true
    added_params["uid"]           =true
    added_params["aid"]           =true
    added_params["authKey"]       =true
    added_params["sessionKey"]    =true
    added_params["clientPlatform"]=false
    added_params["clientVersion"] =false

    for key, required := range added_params {
        val:=input_params[key]
        if ((val==nil || val=="") && required) {
            return nil, errors.New("required ("+key+") param is not found")
        } 
        if (val!=nil && val!="") {
            result_params[key]=fmt.Sprint(val)
        }
    }
    result_params["action"]=action
    result_params["controller"]=controller
    return result_params, nil
}

func  (s signer) calcArraySignMap(params map[string]interface{}) string {
    var result []string
    var tmp string
    for _, key:= range s.getSortedKeys(params) {
        val:=params[key]
        switch concreteVal := val.(type) {
            case map[string]interface{}:
                tmp = s.calcArraySignMap(val.(map[string]interface{}))
            case []interface{}:
                tmp = s.calcArraySignArray(val.([]interface{}))
            default:
                tmp = fmt.Sprint("",concreteVal)
        }
        result = append(result, key+"="+tmp)
    }
    fmt.Println(strings.Join(result,"&"))
    return s.hashMd5(strings.Join(result,"&"))
}

func  (s signer) calcArraySignArray(anArray []interface{}) string {
    var result []string
    var tmp string
    for i, val := range anArray {
        switch concreteVal := val.(type) {
        case map[string]interface{}:
            tmp = s.calcArraySignMap(val.(map[string]interface{}))
        case []interface{}:
            tmp = s.calcArraySignArray(val.([]interface{}))
        default:
            tmp = fmt.Sprint("",concreteVal)
        }
        result=append(result,fmt.Sprint(i)+"="+tmp)
    }
    fmt.Println(strings.Join(result,"&"))
    return s.hashMd5(strings.Join(result,"&"))
}


func  (s signer) getSortedKeys(params map[string]interface{}) []string { 
    var keys []string
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    return keys
}

func  (s signer) hashMd5(input string) string {
    hasher := md5.New()
    hasher.Write([]byte(input))
    return hex.EncodeToString(hasher.Sum(nil))
}

func  (s signer) getControllerActionNames(req *http.Request) (string, string) {
    ps := strings.SplitN(req.URL.Path, "/", 3)
    controller:=ps[1]
    action:= ps[2]
    return controller, action
}

func  (s signer) Execute (res http.ResponseWriter, req *http.Request) ( *http.Request, error) {
    var data map[string]interface{}        
    controller, action:=s.getControllerActionNames(req)
    request_body, _ := ioutil.ReadAll(req.Body)
    if err := json.Unmarshal([]byte(request_body), &data); err != nil {
        panic(err)
    }

    params, _ := s.prepareArray(controller,action,data)
    sign:= s.calcArraySignMap(params)
    if (data["sign"] == sign) {
        return req, nil
    } else {
        return nil, errors.New("request sign is not valid")
    }
}

func (s signer) Init() func( http.ResponseWriter, *http.Request ) ( *http.Request, error ) {
    return s.Execute
}
var Plugin signer
