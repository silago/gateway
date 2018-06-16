package main

import (
	"log"
	"net/http"
	"os"
    //"errors"
	//"github.com/silago/gateway/lib"
	"lib"
    "fmt"
)



/* 
        

        for i in proxy {


        }
        
        createProxy {
            ModifyResponse {
                next_proxy() 
            }
        }
        
*/


func ENV(name string) string {
    result:=""
    if s, ok := os.LookupEnv(name); ok {
        result = s
    } else {
        log.Fatal("Could not get env var " +  name)
    }
    return result
}


func main() {
	var (
		configPath string
		port       string
	)
	middlewares := map[string]func(http.ResponseWriter, *http.Request) ( *http.Request , error) {
        //"auth": TokenAuth,
        "auth": NewAuthenticator(ENV("DB_DRIVER"),ENV("DB_HOST"),ENV("DB_USER"),ENV("DB_PASS"),ENV("DB_NAME"),ENV("DB_CHARSET")).TokenAuth,
        "sign": SignCheck,
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

    fmt.Println("",c)

    http.HandleFunc("/init/", func(w http.ResponseWriter, r *http.Request) {
        r.ParseForm()
        fmt.Fprintf(w, "Hello, %q")
    })

	http.HandleFunc("/", lib.New(c, middlewares))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
