package main

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
)
import b64 "encoding/base64"

var passw = "FOOBAR2451"

type auth struct {
}

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

func (s auth) checkToken(token string) bool {

	/*
		guid := xid.New()
		obj := gabs.New()
		obj.Set(userId, "id")
		obj.Set(guid, "guid")
		return string(encrypt([]byte(obj.String()), passw))
	*/
	return true
}

func (s auth) after(request *http.Request) (*http.Request, error) {
	return request, nil
}

func (s auth) Init() func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	return func(request *http.Request, wrapped func(*http.Request) (*http.Response, error)) (*http.Response, error) {
		response, err := wrapped(request)
		return response, err

	}
}

var Plugin auth
