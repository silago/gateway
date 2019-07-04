package main

import (
	//	"fmt"
	"gateway/lib"
	"log"
	"syscall"

	//	"net"
	"net/http"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var (
		configPath string
	)
	port := ENV("PORT")
	sslPort := ENV("SSL_PORT")
	if len(os.Args) != 2 {
		if s, ok := os.LookupEnv("GATEWAY_CONFIG_FILE"); ok {
			configPath = s
		} else {
			log.Fatal("Usage: gateway path-to-config.json")
		}
	} else {
		configPath = os.Args[1]
	}

	//if config, err := lib.LoadConfig(configPath); err!=nil {
	//	log.Fatal(err)
	//} else {

	gateway := lib.InitGateway(configPath)
	gateway.StartReloadOnSignal(syscall.SIGHUP)
	go gateway.StartTcpPortForwarding()
	http.HandleFunc("/", gateway.GetHandler())
	//lib.InitGateway(configPath).Start()
	http.HandleFunc("/google13dd0d8dae2fd927.html", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("google-site-verification: google13dd0d8dae2fd927.html"))
	})

	log.Println("listen http at ", port)
	go func() {
		log.Fatal(http.ListenAndServeTLS(":"+sslPort, "server.crt", "server.key", nil))
	}()
	log.Fatal(http.ListenAndServe(":"+port, nil))
	select {}
	//}
}
