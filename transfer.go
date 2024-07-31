package wxpayslim

import (
	"context"
	"encoding/xml"
	"time"
)

const (
	transferUrl      = prefix + "/mmpaymkttransfers/promotion/transfers"
	transferQueryUrl = prefix + "/mmpaymkttransfers/gettransferinfo"

	v3TransferUrl = prefix + "/v3/transfer/batches"
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

// Transfer money to user. Docs:
// https://pay.weixin.qq.com/docs/merchant/apis/batch-transfer-to-balance/transfer-batch/initiate-batch-transfer.html
func (client *Client) TransferV3(ctx context.Context, req V3TransferRequests) (*V3TransferResponse, error) {
	var res V3TransferResponse
	if err := client.postJson(ctx, v3TransferUrl, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// V3TransferRequests is used in TransferV3() function.
type V3TransferRequests struct {
	AppId           string              // required
	OutBatchNo      string              // required
	BatchName       string              // required
	BatchRemark     string              // required
	TransferSceneId string              // optional
	NotifyUrl       string              // optional
	Transfers       []V3TransferRequest // required
}

type V3TransferRequest struct {
	OutDetailNo    string
	TransferAmount int
	TransferRemark string
	OpenId         string
	UserName       string
}

var _ jsonRequestable = (*V3TransferRequests)(nil)

func (r V3TransferRequests) toJson(client *Client) requestJson {
	req := transferRequestJson{}
	req.AppId = r.AppId
	req.OutBatchNo = r.OutBatchNo
	req.BatchName = r.BatchName
	req.BatchRemark = r.BatchRemark
	req.TransferSceneId = r.TransferSceneId
	req.NotifyUrl = r.NotifyUrl
	req.TransferDetailList = make([]v3TransferDetail, len(r.Transfers))
	for i := range r.Transfers {
		req.TransferDetailList[i].OutDetailNo = r.Transfers[i].OutDetailNo
		req.TransferDetailList[i].TransferAmount = r.Transfers[i].TransferAmount
		req.TransferDetailList[i].TransferRemark = r.Transfers[i].TransferRemark
		req.TransferDetailList[i].OpenId = r.Transfers[i].OpenId
		req.TransferDetailList[i].UserName = r.Transfers[i].UserName
		req.TotalAmount += r.Transfers[i].TransferAmount
		req.TotalNum += 1
	}
	return req
}

type transferRequestJson struct {
	AppId              string             `json:"appid"`
	OutBatchNo         string             `json:"out_batch_no"`
	BatchName          string             `json:"batch_name"`
	BatchRemark        string             `json:"batch_remark"`
	TotalAmount        int                `json:"total_amount"`
	TotalNum           int                `json:"total_num"`
	TransferDetailList []v3TransferDetail `json:"transfer_detail_list"`
	TransferSceneId    string             `json:"transfer_scene_id,omitempty"`
	NotifyUrl          string             `json:"notify_url,omitempty"`
}

type v3TransferDetail struct {
	OutDetailNo    string `json:"out_detail_no"`
	TransferAmount int    `json:"transfer_amount"`
	TransferRemark string `json:"transfer_remark"`
	OpenId         string `json:"openid"`
	UserName       string `json:"user_name,omitempty"`
}

type V3TransferResponse struct {
	JsonResponse
	OutBatchNo  string    `json:"out_batch_no"`
	BatchId     string    `json:"batch_id"`
	CreateTime  time.Time `json:"create_time"`
	BatchStatus string    `json:"batch_status"`
}

var _ responsible = (*V3TransferResponse)(nil)

func (r V3TransferResponse) AsError() error {
	return JsonResponseError(r.JsonResponse)
}
