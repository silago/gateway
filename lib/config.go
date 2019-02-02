package lib

import (
	"encoding/json"
	"io/ioutil"
)

// Config is a struct that holds the configuration of the gateway.
type Service struct {
	Service   string
	Protocol  string
	Scheme    string
	Plugins   []string
	Aggregate map[string]AggregatePipe
	Chain     []Chain
	Pipes     []Pipe
}

type AggregatePipe struct {
	Service  string
	Endpoint string
	Mapping  map[string]string
}

type Chain struct {
	Service  string
	Endpoint string
}

type Pipe struct {
	Service  string
	Endpoint string
	Map      map[string]string
}

type Config struct {
	Version          string             `json:"version"`
	Scheme           string             `json:"scheme"`
	PortForward		 map[string]string 	`json:"port_forward"`
	Middleware       map[string]string  `json:"middleware"`
	Rules            map[string]Service `json:"rules"`
	NotFoundResponse interface{}        `json:"not_found_error"`
	FallbackRule     string             `json:"fallback_rule"`
}

// Load loads a configuration file and parses it into a Config struct.
func LoadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return load(b)
}

func load(b []byte) (*Config, error) {
	var config Config
	err := json.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
