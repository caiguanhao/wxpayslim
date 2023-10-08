package wxpayslim

import (
	"context"
	"encoding/xml"
)

const (
	createOrderUrl = prefix + "/pay/unifiedorder"
	queryOrderUrl  = prefix + "/pay/orderquery"
)

// CreateOrder initiates payment.
func (client *Client) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	var res CreateOrderResponse
	if err := client.postXml(ctx, createOrderUrl, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type CreateOrderRequest struct {
	AppId          string // required
	DeviceInfo     string // optional
	SignType       string // optional, either MD5 (default) or HMAC-SHA256
	Body           string // required, max length is 127
	Detail         string // optional, max length is 6000
	Attach         string // optional, max length is 127
	OutTradeNo     string // required, max length is 32, min length is 6
	FeeType        string // optional, defaults to CNY
	TotalFee       int    // required, in cents
	SpbillCreateIp string // required, user's ip address
	TimeStart      string // optional, UTC+8 time format: 20060102150405
	TimeExpire     string // optional, UTC+8 time format: 20060102150405
	GoodsTag       string // optional, max length is 32
	NotifyURL      string // required, max length is 256
	TradeType      string // required, can be JSAPI, NATIVE, APP
	ProductId      string // required if TradeType == NATIVE, max length is 32
	LimitPay       string // optional, set to no_credit to disallow credit cards
	OpenId         string // required if TradeType == JSAPI
	Receipt        string // optional, set to Y to enable receipt
	ProfitSharing  string // optional, either Y or N (default)
	SceneInfo      string // optional
}

var _ requestable = (*CreateOrderRequest)(nil)

func (r CreateOrderRequest) toXml(client *Client) requestXml {
	req := createOrderRequestXml{}
	copyFields(r, &req)
	req.MchId = client.MchId
	req.NonceStr = randomStr(32)
	req.Sign = client.generateSign(req)
	return req
}

type createOrderRequestXml struct {
	XMLName        xml.Name `xml:"xml"`
	AppId          string   `xml:"appid"`
	MchId          string   `xml:"mch_id"`
	DeviceInfo     string   `xml:"device_info,omitempty"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	SignType       string   `xml:"sign_type,omitempty"`
	Body           string   `xml:"body"`
	Detail         string   `xml:"detail,omitempty"`
	Attach         string   `xml:"attach,omitempty"`
	OutTradeNo     string   `xml:"out_trade_no"`
	FeeType        string   `xml:"fee_type,omitempty"`
	TotalFee       int      `xml:"total_fee"`
	SpbillCreateIp string   `xml:"spbill_create_ip"`
	TimeStart      string   `xml:"time_start,omitempty"`
	TimeExpire     string   `xml:"time_expire,omitempty"`
	GoodsTag       string   `xml:"goods_tag,omitempty"`
	NotifyURL      string   `xml:"notify_url"`
	TradeType      string   `xml:"trade_type"`
	ProductId      string   `xml:"product_id,omitempty"`
	LimitPay       string   `xml:"limit_pay,omitempty"`
	OpenId         string   `xml:"openid,omitempty"`
	Receipt        string   `xml:"receipt,omitempty"`
	ProfitSharing  string   `xml:"profit_sharing,omitempty"`
	SceneInfo      string   `xml:"scene_info,omitempty"`
}

type CreateOrderResponse struct {
	Response
	AppId      string `xml:"mch_appid,omitempty"`
	MchId      string `xml:"mchid,omitempty"`
	DeviceInfo string `xml:"device_info,omitempty"`
	TradeType  string `xml:"trade_type"`
	PrepayId   string `xml:"prepay_id"`
	CodeUrl    string `xml:"code_url"`
}

var _ responsible = (*CreateOrderResponse)(nil)

func (r CreateOrderResponse) AsError() error {
	return ResponseError(r.Response)
}

// QueryOrder gets information of an order by Transaction ID or Trade No.
func (client *Client) QueryOrder(ctx context.Context, req QueryOrderRequest) (*QueryOrderResponse, error) {
	var res QueryOrderResponse
	if err := client.postXml(ctx, queryOrderUrl, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type QueryOrderRequest struct {
	AppId         string // required
	TransactionId string // either TransactionId or OutTradeNo is required
	OutTradeNo    string
}

var _ requestable = (*QueryOrderRequest)(nil)

func (r QueryOrderRequest) toXml(client *Client) requestXml {
	req := queryOrderRequestXml{}
	copyFields(r, &req)
	req.MchId = client.MchId
	req.NonceStr = randomStr(32)
	req.Sign = client.generateSign(req)
	return req
}

type queryOrderRequestXml struct {
	XMLName       xml.Name `xml:"xml"`
	AppId         string   `xml:"appid"`
	MchId         string   `xml:"mch_id"`
	TransactionId string   `xml:"transaction_id,omitempty"`
	OutTradeNo    string   `xml:"out_trade_no,omitempty"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	SignType      string   `xml:"sign_type,omitempty"`
}

type QueryOrderResponse struct {
	Response
	AppId string `xml:"appid,omitempty"`
	MchId string `xml:"mch_id,omitempty"`

	DeviceInfo         string `xml:"device_info,omitempty"`
	OpenId             string `xml:"openid"`
	IsSubscribe        string `xml:"is_subscribe"`
	TradeType          string `xml:"trade_type"`
	TradeState         string `xml:"trade_state"`
	BankType           string `xml:"bank_type"`
	TotalFee           int    `xml:"total_fee"`
	SettlementTotalFee int    `xml:"settlement_total_fee"`
	FeeType            string `xml:"fee_type"`
	CashFee            int    `xml:"cash_fee"`
	CashFeeType        string `xml:"cash_fee_type"`
	CouponFee          int    `xml:"coupon_fee"`
	CouponCount        int    `xml:"coupon_count"`
	TransactionId      string `xml:"transaction_id"`
	OutTradeNo         string `xml:"out_trade_no"`
	Attach             string `xml:"attach"`
	TimeEnd            string `xml:"time_end"`
	TradeStateDesc     string `xml:"trade_state_desc"`
}

var _ responsible = (*QueryOrderResponse)(nil)

func (r QueryOrderResponse) AsError() error {
	return ResponseError(r.Response)
}

// Check if order is successfully paid.
func (r QueryOrderResponse) Paid() bool {
	return r.ReturnCode == "SUCCESS" && r.ResultCode == "SUCCESS" && r.TradeState == "SUCCESS"
}
