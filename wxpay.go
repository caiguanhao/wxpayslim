package wxpayslim

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"math/rand"
	"sort"
	"time"
)

const prefix = "https://api.mch.weixin.qq.com"

type Client struct {
	MchId string
	Key   string
	Debug bool

	TLSClientConfig *tls.Config
}

// NewClient creates a new client.
func NewClient(mchId, key string) *Client {
	return &Client{
		MchId: mchId,
		Key:   key,
	}
}

// Set certificate (apiclient_cert.pem, string starts with -----BEGIN
// CERTIFICATE-----) and private key (apiclient_key.pem, string starts with
// -----BEGIN PRIVATE KEY-----). If you have different certificate format, set
// client's TLSClientConfig property directly.
func (client *Client) SetCertificate(certPEM, keyPem string) error {
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPem))
	if err != nil {
		return err
	}
	client.TLSClientConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return nil
}

type Response struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
	NonceStr   string `xml:"nonce_str,omitempty"`
	ResultCode string `xml:"result_code,omitempty"`
	ErrCode    string `xml:"err_code,omitempty"`
	ErrCodeDes string `xml:"err_code_des,omitempty"`
}

func (r Response) Success() bool {
	return r.ReturnCode == "SUCCESS" && r.ResultCode == "SUCCESS"
}

type ResponseError Response

func (r ResponseError) Error() string {
	if r.ErrCode == "" {
		return "UNKNOWN"
	}
	return r.ErrCode + ": " + r.ErrCodeDes
}

func paramsToString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	for _, k := range keys {
		if params[k] == "" {
			continue
		}
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(params[k])
	}
	return buf.String()
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

type Utc8Time time.Time

func (tm Utc8Time) String() string {
	return time.Time(tm).String()
}

func (tm *Utc8Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	err := d.DecodeElement(&value, &start)
	if err != nil {
		return err
	}
	loc := time.FixedZone("UTC+8", 8*60*60)
	t, err := time.ParseInLocation("2006-01-02 15:04:05", value, loc)
	if err != nil {
		return err
	}
	*tm = Utc8Time(t)
	return nil
}
