package main

import (
    "encoding/base64" 
    "errors" 
    "fmt" 
    "strings" 
    "crypto/sha256" 
    "net/http" 
)

type PerfTokenMiddleware struct {
    salt string
}

func NewPerfToken (salt string ) PerfTokenMiddleware {
   var auth PerfTokenMiddleware
   return auth
} 


func (p *PerfTokenMiddleware) checkAccessToken(token string) (string, string, error) {
    str, err := base64.StdEncoding.DecodeString(token)
    if ( err!= nil ) {
        return "","", err
    }
    splitted:= strings.Split(string(str[:]),":")
    if (len(splitted) !=3 ) {
        return "", "", errors.New("token is not valid")
    } 
    platform_name:=splitted[0]
    user_id:=splitted[1]
    hash:=splitted[2]
    if !p.checkHash(hash, platform_name, user_id) {
        return "", "", errors.New("Hash is not valid")
    }
    return user_id, platform_name, nil
}

func (p *PerfTokenMiddleware) checkHash(hash string, platform_name string, user_id string) bool  {
    return hash == p.generateHash(platform_name, user_id) 
}

func (p *PerfTokenMiddleware) initFromEnv()  *PerfTokenMiddleware {
    salt:= ENV("PERF_SALT")
    p.salt = salt
    return p
}

func (p *PerfTokenMiddleware) generateHash(platform_name string, user_id string) string {
     str:=platform_name + "$" + p.salt + "$" + user_id + "$" + p.salt
     result:= sha256.New().Sum([]byte(str))
     return string(result[:])
}

/* 
 * todo:: describe forward rule
 */   
func (p *PerfTokenMiddleware) TokenAuth(res http.ResponseWriter, req *http.Request) ( *http.Request, error) {
        req.ParseForm()
        token:=req.Form.Get("auth_token")
        if (token == "" ) {
            return nil, errors.New("Auth token is empty" );
        }
        user_id, platform_name, err :=p.checkAccessToken(token);

        if err!=nil {
            req.Form.Add("user_id", fmt.Sprint(user_id))
            req.Form.Add("platform_name", fmt.Sprint(platform_name))
            return nil, errors.New("Auth token is not valid" );
        } else {
            return req, nil
        }
}
