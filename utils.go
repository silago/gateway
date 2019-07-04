package main

import "os"
import "log"

func ENV(name string) string {
	result := ""
	if s, ok := os.LookupEnv(name); ok {
		result = s
	} else {
		log.Fatal("Could not get env var " + name)
	}
	return result
}
