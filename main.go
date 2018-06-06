package main

import (
	"log"
	"net/http"
	"os"
    "errors"

	//"github.com/silago/gateway/lib"
	"lib"
    "fmt"
)


func tokenAuth(res http.ResponseWriter, req *http.Request) ( *http.Request, error) {
    //token:=req.URL.Query().Get("auth_token");
    /* query to db end encodeToken */
    if (false) {
        return nil, errors.New("Auth token is not valid" );
    }
    return req, nil
}

func main() {
	var (
		configPath string
		port       string
	)
	middlewares := map[string]func(http.ResponseWriter, *http.Request) ( *http.Request , error) {
        "auth": tokenAuth,
    }

	if len(os.Args) != 2 {
		if s, ok := os.LookupEnv("GATEWAY_CONFIG_FILE"); ok {
			configPath = s
		} else {
			log.Fatal("Usage: gateway path-to-config.json")
		}
	} else {
		configPath = os.Args[1]
	}

	c, err := lib.Load(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if c.Port == "" {
		p, ok := os.LookupEnv("HTTP_PLATFORM_PORT")
		if !ok {
			log.Fatal("Config file should specify port, or the HTTP_PLATFORM_PORT environment variable must be set.")
		}
		port = p
	} else {
		port = c.Port
	}

	http.HandleFunc("/", lib.New(c, middlewares))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
