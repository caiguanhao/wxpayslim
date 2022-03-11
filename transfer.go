package wxpayslim

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
)

const transferUrl = prefix + "/mmpaymkttransfers/promotion/transfers"

// Transfer money to user. Docs:
// https://pay.weixin.qq.com/wiki/doc/api/tools/mch_pay.php?chapter=14_2
func (client *Client) Transfer(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	xmlData, err := xml.MarshalIndent(req.toXml(client), "", "  ")
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", transferUrl, bytes.NewBuffer(xmlData))
	if err != nil {
		return nil, err
	}
	if client.Debug {
		dump, err := httputil.DumpRequestOut(httpReq, true)
		if err != nil {
			return nil, err
		}
		log.Println(string(dump))
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: client.TLSClientConfig,
		},
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if client.Debug {
		dumpBody := strings.Contains(resp.Header.Get("Content-Type"), "text/")
		dump, err := httputil.DumpResponse(resp, dumpBody)
		if err != nil {
			return nil, err
		}
		log.Println(string(dump))
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res TransferResponse
	err = xml.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}
	if res.Success() {
		return &res, nil
	} else {
		return nil, ResponseError(res.Response)
	}
}

// TransferRequest is used in Transfer() function.
type TransferRequest struct {
	AppId          string // required
	OpenId         string // required
	DeviceInfo     string // optional
	PartnerTradeNo string // required
	CheckName      string // optional, either NO_CHECK (default) or FORCE_CHECK
	ReUserName     string // required if CheckName is FORCE_CHECK
	Amount         int    // required, must not lower than 100 (1.00 yuan)
	Desc           string // required
	SpbillCreateIp string // optional, user's IP address
}

func (r TransferRequest) toXml(client *Client) transferRequestXml {
	checkName := r.CheckName
	if checkName == "" {
		checkName = "NO_CHECK"
	}
	req := transferRequestXml{
		AppId:          r.AppId,
		MchId:          client.MchId,
		DeviceInfo:     r.DeviceInfo,
		NonceStr:       randomStr(32),
		PartnerTradeNo: r.PartnerTradeNo,
		OpenId:         r.OpenId,
		CheckName:      checkName,
		ReUserName:     r.ReUserName,
		Amount:         r.Amount,
		Desc:           r.Desc,
		SpbillCreateIp: r.SpbillCreateIp,
	}
	req.Sign = req.generateSign(client)
	return req
}

type transferRequestXml struct {
	XMLName        xml.Name `xml:"xml"`
	AppId          string   `xml:"mch_appid"`
	MchId          string   `xml:"mchid"`
	DeviceInfo     string   `xml:"device_info,omitempty"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	PartnerTradeNo string   `xml:"partner_trade_no"`
	OpenId         string   `xml:"openid"`
	CheckName      string   `xml:"check_name"`
	ReUserName     string   `xml:"re_user_name,omitempty"`
	Amount         int      `xml:"amount"`
	Desc           string   `xml:"desc"`
	SpbillCreateIp string   `xml:"spbill_create_ip,omitempty"`
}

func (r transferRequestXml) generateSign(client *Client) string {
	params := map[string]string{
		"mch_appid":        r.AppId,
		"mchid":            r.MchId,
		"device_info":      r.DeviceInfo,
		"nonce_str":        r.NonceStr,
		"partner_trade_no": r.PartnerTradeNo,
		"openid":           r.OpenId,
		"check_name":       r.CheckName,
		"re_user_name":     r.ReUserName,
		"amount":           strconv.Itoa(r.Amount),
		"desc":             r.Desc,
		"spbill_create_ip": r.SpbillCreateIp,
	}
	str := paramsToString(params) + "&key=" + client.Key
	return fmt.Sprintf("%X", md5.Sum([]byte(str)))
}

type TransferResponse struct {
	Response
	AppId          string    `xml:"mch_appid,omitempty"`
	MchId          string    `xml:"mchid,omitempty"`
	DeviceInfo     string    `xml:"device_info,omitempty"`
	PartnerTradeNo string    `xml:"partner_trade_no"`
	PaymentNo      string    `xml:"payment_no"`
	PaymentTime    *Utc8Time `xml:"payment_time"`
}
