package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/rs/xid"

	b64 "encoding/base64"
)

const authTokenHeader = "X-Auth-Token"
var password = os.Getenv("AUTH_PLUGIN_PASSWORD")
type auth struct {}

func xor(input, key string) (output string) {
	for i := 0; i < len(input); i++ {
		output += string(input[i] ^ key[i%len(key)])
	}

	return output
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
func encrypt(data string, passphrase string) string {
	return b64.StdEncoding.EncodeToString([]byte(xor(data, passphrase)))
}

func decrypt(data string, passphrase string) string {
	sDec, _ := b64.StdEncoding.DecodeString(xor(data, passphrase))
	return string(sDec)
}
func (s auth) generateToken(userId string) string {
	guid := xid.New()
	obj := gabs.New()
	_, _ = obj.Set(userId, "id")
	_, _ = obj.Set(guid, "guid")
	return string(encrypt(obj.String(), password))
}

func (s auth) after(response *http.Response) (*http.Response, error) {

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		response.Body = ioutil.NopCloser(strings.NewReader(fmt.Sprint("", err)))
		return response, nil
	}

	obj, err := gabs.ParseJSON(body)
	if err != nil {
		response.Body = ioutil.NopCloser(strings.NewReader(err.Error()))
		return response, nil
	}

	value := obj.Path("id").String()
	if value != "" {
		token := s.generateToken(value)
		response.Header.Set(authTokenHeader, token)
		fmt.Println("token been set")
	} else {
		fmt.Println("could not parse id in responce: ", obj.String())
	}

	response.Body = ioutil.NopCloser(strings.NewReader(obj.String()))

	return response, nil
}

func (s auth) Init() func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	if password == "" {
		panic("AUTH_PLUGIN_PASSWORD is not set")
	}
	return func(request *http.Request, wrapped func(*http.Request) (*http.Response, error)) (*http.Response, error) {
		response, _ := wrapped(request)
		return s.after(response)
	}
}
/* Plugin: generates auth-token and adds it to the headers */
var Plugin auth
