package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/caiguanhao/wxpayslim"
)

func main() {
	mchid := flag.String("mchid", "", "merchant's id")
	key := flag.String("key", "", "merchant's key")
	appid := flag.String("appid", "", "app id")
	body := flag.String("body", "TEST", "payment body")
	fee := flag.Int("fee", 1, "payment total fee")
	outTradeNo := flag.String("no", "", "out trade no, random string if empty")
	notifyUrl := flag.String("url", "http://localhost/", "notify url")
	flag.Parse()

	if *outTradeNo == "" {
		no := randomStr(20)
		outTradeNo = &no
	}

	client := wxpayslim.NewClient(*mchid, *key)
	resp, err := client.CreateOrder(context.Background(), wxpayslim.CreateOrderRequest{
		AppId:          *appid,
		Body:           *body,
		OutTradeNo:     *outTradeNo,
		TotalFee:       *fee,
		SpbillCreateIp: "127.0.0.1",
		NotifyURL:      *notifyUrl,
		TradeType:      "NATIVE",
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(resp.CodeUrl)
}

func randomStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
