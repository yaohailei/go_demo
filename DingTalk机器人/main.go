package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/imroc/req"
)

// Secret 加密密钥
var Secret = "SEC4a2febcd77003fb58f1b86f7d0d4d92e43a2801eaa45ec7fb76bc5f103c55401"

// Webhook 地址
var Webhook = "https://oapi.dingtalk.com/robot/send?access_token=fd16256e73f44a9da82f926dc8ba4a52f65df2c84b22b1742bb2a2ed8d28a0a5"

// ResponseData 定义钉钉返回信息结构体
type ResponseData struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// TextInfo 定义发送text
type TextInfo struct {
	At      At     `json:"at"`
	Text    Text   `json:"text"`
	Msgtype string `json:"msgtype"`
}
type At struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}
type Text struct {
	Content string `json:"content"`
}

// 定义发送信息方法
func sendMessage(url string, timestamped string, sign string, content string, phone ...string) {
	param := req.Param{
		"timestamp": timestamped,
		"sign":      sign,
	}
	// 获取被@人手机，存入数组
	var phoneList []string
	for _, data := range phone {
		phoneList = append(phoneList, data)
	}

	// 定义发送text信息
	msg := TextInfo{
		At: At{
			AtMobiles: phoneList,
			IsAtAll:   false,
		},
		Text:    Text{Content: content},
		Msgtype: "text",
	}
	// 发送请求
	r, errRequest := req.Post(url, req.BodyJSON(&msg), param)
	responses := ResponseData{}
	if errRequest != nil {
		log.Fatal(errRequest)
	}
	_ = r.ToJSON(&responses)
	log.Printf("Errcode: %d ,Errmsg: %s", responses.Errcode, responses.Errmsg)
}

//获取签名方法
func getSign(timestamp int64, secret string) (string, string) {
	timestamped := strconv.FormatInt(timestamp, 10)
	//拼接签名需要加密字符串
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	//HmacSHA256算法计算签名
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(stringToSign))
	sha := hmac256.Sum(nil)
	// 返回字符串类型时间戳和签名
	return timestamped, base64.StdEncoding.EncodeToString(sha)
}

func main() {
	// 获取当前时间
	timestamp := time.Now().UnixNano() / 1e6
	// 加签
	timestamped, sign := getSign(timestamp, Secret)
	// 发送消息
	sendMessage(Webhook, timestamped, sign, "Hello World!", "13372024567", "18061931234")
}
