package wxpayslim

import (
	"context"
	"encoding/xml"
)

const (
	transferUrl      = prefix + "/mmpaymkttransfers/promotion/transfers"
	transferQueryUrl = prefix + "/mmpaymkttransfers/gettransferinfo"
)

// Transfer money to user. Docs:
// https://pay.weixin.qq.com/wiki/doc/api/tools/mch_pay.php?chapter=14_2
func (client *Client) Transfer(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	var res TransferResponse
	if err := client.postXml(ctx, transferUrl, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
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

var _ requestable = (*TransferRequest)(nil)

func (r TransferRequest) toXml(client *Client) requestXml {
	req := transferRequestXml{}
	copyFields(r, &req)
	req.MchId = client.MchId
	req.NonceStr = randomStr(32)
	checkName := r.CheckName
	if checkName == "" {
		checkName = "NO_CHECK"
	}
	req.CheckName = checkName
	req.Sign = client.generateSign(req)
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

type TransferResponse struct {
	Response
	AppId          string    `xml:"mch_appid,omitempty"`
	MchId          string    `xml:"mchid,omitempty"`
	DeviceInfo     string    `xml:"device_info,omitempty"`
	PartnerTradeNo string    `xml:"partner_trade_no"`
	PaymentNo      string    `xml:"payment_no"`
	PaymentTime    *Utc8Time `xml:"payment_time"`
}

var _ responsible = (*TransferResponse)(nil)

func (r TransferResponse) AsError() error {
	return ResponseError(r.Response)
}

// Query existing transaction. Docs:
// https://pay.weixin.qq.com/wiki/doc/api/tools/mch_pay.php?chapter=14_3
func (client *Client) TransferQuery(ctx context.Context, req TransferQueryRequest) (*TransferQueryResponse, error) {
	var res TransferQueryResponse
	if err := client.postXml(ctx, transferQueryUrl, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// TransferQueryRequest is used in TransferQuery() function.
type TransferQueryRequest struct {
	AppId          string // required
	PartnerTradeNo string // required
}

var _ requestable = (*TransferQueryRequest)(nil)

func (r TransferQueryRequest) toXml(client *Client) requestXml {
	req := transferQueryRequestXml{}
	copyFields(r, &req)
	req.MchId = client.MchId
	req.NonceStr = randomStr(32)
	req.Sign = client.generateSign(req)
	return req
}

type transferQueryRequestXml struct {
	XMLName        xml.Name `xml:"xml"`
	AppId          string   `xml:"appid"`
	MchId          string   `xml:"mch_id"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	PartnerTradeNo string   `xml:"partner_trade_no"`
}

type TransferQueryResponse struct {
	Response
	AppId          string    `xml:"appid,omitempty"`
	MchId          string    `xml:"mch_id,omitempty"`
	DetailId       string    `xml:"detail_id,omitempty"`
	Status         string    `xml:"status,omitempty"`
	Reason         string    `xml:"reason,omitempty"`
	OpenId         string    `xml:"openid,omitempty"`
	TransferName   string    `xml:"transfer_name,omitempty"`
	PartnerTradeNo string    `xml:"partner_trade_no"`
	PaymentAmount  int       `xml:"payment_amount"`
	TransferTime   *Utc8Time `xml:"transfer_time"`
	PaymentTime    *Utc8Time `xml:"payment_time"`
	Desc           string    `xml:"desc"`
}

var _ responsible = (*TransferQueryResponse)(nil)

func (r TransferQueryResponse) AsError() error {
	return ResponseError(r.Response)
}
