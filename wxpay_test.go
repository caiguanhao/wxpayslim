package wxpayslim

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"
)

type clientConfig struct {
	Mchid   string
	Key     string
	Cert    string
	Certkey string
	Appid   string
	Openid  string
}

var client *Client
var config *clientConfig

func init() {
	// test.json should look like this:
	// {
	// 	"mchid": "1111111111",
	// 	"key": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	// 	"cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
	// 	"certkey": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
	//	"appid": "wxxxxxxxxxxxxxxxxx",
	//	"openid": "oAxxxxxxxxxxxxxxxxxxxxxxxxxx"
	// }
	b, err := os.ReadFile("test.json")
	if err != nil {
		log.Println(err)
		return
	}
	config = new(clientConfig)
	err = json.Unmarshal(b, config)
	if err != nil {
		log.Println(err)
		config = nil
		return
	}
	client = NewClient(config.Mchid, config.Key)
	client.Debug = os.Getenv("DEBUG") == "1"
	err = client.SetCertificate(config.Cert, config.Certkey)
	if err != nil {
		log.Println(err)
		client = nil
		return
	}
}

func TestCreateOrder(t *testing.T) {
	ctx := context.Background()
	resp, err := client.CreateOrder(ctx, CreateOrderRequest{
		AppId:          config.Appid,
		Body:           "test",
		OutTradeNo:     randomStr(32),
		TotalFee:       1,
		SpbillCreateIp: "127.0.0.1",
		NotifyURL:      "http://localhost/",
		TradeType:      "NATIVE",
		ProductId:      randomStr(32),
	})
	if err != nil {
		t.Error("expected error to be nil:", err)
	}
	t.Log("qrcode url", resp.CodeUrl)
}

func TestTransfer(t *testing.T) {
	if client == nil {
		t.Log("client is not initialized, skipped")
		return
	}
	tradeNo := "TESTz20220311z111122" // "TESTz" + time.Now().Format("20060102z150405")
	ctx := context.Background()
	_, err := client.Transfer(ctx, TransferRequest{
		AppId:          config.Appid,
		PartnerTradeNo: tradeNo,
		OpenId:         config.Openid,
		Amount:         1,
		Desc:           "测试",
	})
	if err == nil {
		t.Error("expected error to be not nil")
	} else {
		expected := "FATAL_ERROR: 更换了金额，但商户单号未更新"
		if err.Error() != expected {
			t.Error("expected error to be:", expected)
		} else {
			t.Log("error test passed")
		}
	}
	resp, err := client.Transfer(ctx, TransferRequest{
		AppId:          config.Appid,
		PartnerTradeNo: tradeNo,
		OpenId:         config.Openid,
		Amount:         100,
		Desc:           "测试",
	})
	if err != nil {
		t.Error("expected error to be nil:", err)
	} else {
		t.Logf("%+v", resp)
		if resp.Success() != true {
			t.Error("expected resp to be success")
		} else {
			t.Log("Transfer() test passed")
		}
	}

	resp2, err := client.TransferQuery(ctx, TransferQueryRequest{
		AppId:          config.Appid,
		PartnerTradeNo: tradeNo,
	})
	if err != nil {
		t.Error("expected error to be nil:", err)
	} else {
		t.Logf("%+v", resp2)
		if resp2.Success() != true {
			t.Error("expected resp to be success")
		} else {
			t.Log("TransferQuery() test passed")
		}
	}
}
