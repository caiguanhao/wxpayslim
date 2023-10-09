package wxpayslim

import (
	"context"
	"encoding/xml"
)

const (
	createOrderUrl = prefix + "/pay/unifiedorder"
	queryOrderUrl  = prefix + "/pay/orderquery"
	refundOrderUrl = prefix + "/secapi/pay/refund"
	queryRefundUrl = prefix + "/pay/refundquery"
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

// RefundOrder initiates refund. Need to set certificate (client.SetCertificate) first.
func (client *Client) RefundOrder(ctx context.Context, req RefundOrderRequest) (*RefundOrderResponse, error) {
	var res RefundOrderResponse
	if err := client.postXml(ctx, refundOrderUrl, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type RefundOrderRequest struct {
	AppId         string // required
	SignType      string // optional, either MD5 (default) or HMAC-SHA256
	TransactionId string // either TransactionId or OutTradeNo is required
	OutTradeNo    string
	OutRefundNo   string // required, max length is 64
	TotalFee      int    // required, in cents
	RefundFee     int    // required, in cents
	RefundFeeType string // optional, defaults to CNY
	RefundDesc    string // optional
	RefundAccount string // optional
	NotifyURL     string // optional
}

var _ requestable = (*RefundOrderRequest)(nil)

func (r RefundOrderRequest) toXml(client *Client) requestXml {
	req := refundOrderRequestXml{}
	copyFields(r, &req)
	req.MchId = client.MchId
	req.NonceStr = randomStr(32)
	req.Sign = client.generateSign(req)
	return req
}

type refundOrderRequestXml struct {
	XMLName       xml.Name `xml:"xml"`
	AppId         string   `xml:"appid"`
	MchId         string   `xml:"mch_id"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	SignType      string   `xml:"sign_type,omitempty"`
	TransactionId string   `xml:"transaction_id,omitempty"`
	OutTradeNo    string   `xml:"out_trade_no,omitempty"`
	OutRefundNo   string   `xml:"out_refund_no"`
	TotalFee      int      `xml:"total_fee"`
	RefundFee     int      `xml:"refund_fee"`
	RefundFeeType string   `xml:"refund_fee_type,omitempty"`
	RefundDesc    string   `xml:"refund_desc,omitempty"`
	RefundAccount string   `xml:"refund_account,omitempty"`
	NotifyURL     string   `xml:"notify_url,omitempty"`
}

type RefundOrderResponse struct {
	Response
	AppId               string `xml:"appid,omitempty"`
	MchId               string `xml:"mch_id,omitempty"`
	TransactionId       string `xml:"transaction_id"`
	OutTradeNo          string `xml:"out_trade_no"`
	OutRefundNo         string `xml:"out_refund_no"`
	RefundId            string `xml:"refund_id"`
	RefundFee           int    `xml:"refund_fee"`
	SettlementRefundFee int    `xml:"settlement_refund_fee"`
	TotalFee            int    `xml:"total_fee"`
	SettlementTotalFee  int    `xml:"settlement_total_fee"`
	FeeType             string `xml:"fee_type"`
	CashFee             int    `xml:"cash_fee"`
	CashFeeType         string `xml:"cash_fee_type"`
	CashRefundFee       int    `xml:"cash_refund_fee"`
}

var _ responsible = (*RefundOrderResponse)(nil)

func (r RefundOrderResponse) AsError() error {
	return ResponseError(r.Response)
}

// QueryRefundOrder gets information of a refund order by Transaction ID or Trade No.
func (client *Client) QueryRefundOrder(ctx context.Context, req QueryRefundOrderRequest) (*QueryRefundOrderResponse, error) {
	var res QueryRefundOrderResponse
	if err := client.postXml(ctx, queryRefundUrl, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type QueryRefundOrderRequest struct {
	AppId         string // required
	TransactionId string // either TransactionId, OutTradeNo, OutRefundNo or RefundId is required
	OutTradeNo    string
	OutRefundNo   string
	RefundId      string
	Offset        int // optional
}

var _ requestable = (*QueryRefundOrderRequest)(nil)

func (r QueryRefundOrderRequest) toXml(client *Client) requestXml {
	req := queryRefundOrderRequestXml{}
	copyFields(r, &req)
	req.MchId = client.MchId
	req.NonceStr = randomStr(32)
	req.Sign = client.generateSign(req)
	return req
}

type queryRefundOrderRequestXml struct {
	XMLName       xml.Name `xml:"xml"`
	AppId         string   `xml:"appid"`
	MchId         string   `xml:"mch_id"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	SignType      string   `xml:"sign_type,omitempty"`
	TransactionId string   `xml:"transaction_id,omitempty"`
	OutTradeNo    string   `xml:"out_trade_no,omitempty"`
	OutRefundNo   string   `xml:"out_refund_no,omitempty"`
	RefundId      string   `xml:"refund_id,omitempty"`
	Offset        int      `xml:"offset,omitempty"`
}

type QueryRefundOrderResponse struct {
	Response
	AppId string `xml:"appid"`
	MchId string `xml:"mch_id"`

	TotalRefundCount     int    `xml:"total_refund_count"`
	TransactionId        string `xml:"transaction_id"`
	OutTradeNo           string `xml:"out_trade_no"`
	TotalFee             int    `xml:"total_fee"`
	SettlementTotalFee   int    `xml:"settlement_total_fee"`
	FeeType              string `xml:"fee_type,omitempty"`
	CashFee              int    `xml:"cash_fee"`
	RefundCount          int    `xml:"refund_count"`
	OutRefundNo0         string `xml:"out_refund_no_0"`
	RefundId0            string `xml:"refund_id_0"`
	RefundChannel0       string `xml:"refund_channel_0"`
	RefundFee0           int    `xml:"refund_fee_0"`
	RefundFee            int    `xml:"refund_fee"`
	CouponRefundFee      int    `xml:"coupon_refund_fee"`
	SettlementRefundFee0 int    `xml:"settlement_refund_fee_0"`
	RefundStatus0        string `xml:"refund_status_0"`
	RefundAccount0       string `xml:"refund_account_0"`
	RefundRecvAccout0    string `xml:"refund_recv_accout_0"`
	RefundSuccessTime0   string `xml:"refund_success_time_0"`
	CashRefundFee        int    `xml:"cash_refund_fee"`
}

var _ responsible = (*QueryRefundOrderResponse)(nil)

func (r QueryRefundOrderResponse) AsError() error {
	return ResponseError(r.Response)
}

// Check if order is successfully refunded.
func (r QueryRefundOrderResponse) Refunded() bool {
	return r.ReturnCode == "SUCCESS" && r.ResultCode == "SUCCESS" && r.RefundStatus0 == "SUCCESS"
}
