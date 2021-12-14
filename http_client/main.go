package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/imroc/req"
)

//
type Response struct {
	Msg  string        `json:"msg"`
	Code int64         `json:"code"`
	Data []interface{} `json:"data"`
}

func main() {
	const (
		url    = "https://bianque.ssl.ysten.com/yst-bianque/alert/ngkeeper/zhejiang"
		secret = "IeFeTcep16r5dXFdmk_"
	)
	sEnc := base64.StdEncoding.EncodeToString([]byte(secret))

	//设置请求头，传入token
	authHeader := req.Header{
		"Content-Type":   "application/json",
		"Authorizations": sEnc,
	}
	//跳过 SSL 认证，使用非安全传输
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	r, err := req.Get(url, client, authHeader)
	if err != nil {
		fmt.Println(err)
	}
	response := r.Response()
	body, err := ioutil.ReadAll(response.Body)
	//fmt.Println(string(body))
	if err != nil {
		fmt.Println("ioutil.ReadAll error")
	}

	re := Response{}
	if err = json.Unmarshal(body, &re); err == nil {
		fmt.Printf("%++v\n", re)
		//fmt.Println(dat["status"])
	} else {
		fmt.Println(err)
	}
}
