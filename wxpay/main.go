package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
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
	isQuery := flag.Bool("query", false, "query (instead of create) order")
	flag.Parse()

	if *outTradeNo == "" {
		no := randomStr(20)
		outTradeNo = &no
	}

	client := wxpayslim.NewClient(*mchid, *key)
	fmt.Fprintln(os.Stderr, "OutTradeNo:", *outTradeNo)
	if *isQuery {
		resp, err := client.QueryOrder(context.Background(), wxpayslim.QueryOrderRequest{
			AppId:      *appid,
			OutTradeNo: *outTradeNo,
		})
		if err != nil {
			log.Fatalln(err)
		}
		json.NewEncoder(os.Stderr).Encode(resp)
		if resp.TradeState != "SUCCESS" {
			os.Exit(1)
		}
		return
	}
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
