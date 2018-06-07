package main

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
    "fmt"
	"net/http"
	"errors"
)

type User struct {
  UserId uint
  Token string
}

type Authenticator struct {
    Db *gorm.DB
}

func (a *Authenticator) GetUserIDByToken (token string) ( uint, bool )  {
  var record User
  not_found_error:= a.Db.First(&record, "token = ?", token).RecordNotFound()
  if (not_found_error) {
    return 0, false     
  }
  
  return record.UserId, true
}

func NewAuthenticator( db_driver string, db_host string, db_user string, db_pass string, db_name string, db_charset string )  Authenticator {
    var auth Authenticator 
    connection_string:= fmt.Sprintf("%s:%s@%s/%s?charset=%s", db_user, db_pass, db_host, db_name,db_charset)
    db, error := gorm.Open(db_driver, connection_string)
    if (error!=nil) {
        panic(error.Error())
    }
    auth.Db = db
    return auth
}

func (a Authenticator) TokenAuth(res http.ResponseWriter, req *http.Request) ( *http.Request, error) {
    req.ParseForm()
    token:=req.Form.Get("auth_token")
    if (token == "" ) {
        return nil, errors.New("Auth token is empty" );
    }
    if user_id, ok :=a.GetUserIDByToken(token); ok {
        req.Form.Add("user_id", fmt.Sprint(user_id))
        return req, nil
    } else {
        return nil, errors.New("Auth token is not valid" );
    }
}
