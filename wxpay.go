package wxpayslim

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const prefix = "https://api.mch.weixin.qq.com"

const applicationJson = "application/json"

type Client struct {
	MchId string
	Key   string
	Debug bool

	TLSClientConfig  *tls.Config
	certSerialNumber string
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
//
// If you have p12 file (apiclient_cert.p12), you can use following command to
// get certificate and private key:
//
//	openssl pkcs12 -in apiclient_cert.p12 -nodes
func (client *Client) SetCertificate(certPEM, keyPem string) error {
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPem))
	if err != nil {
		return err
	}
	x509cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}
	client.TLSClientConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	client.certSerialNumber = strings.ToUpper(x509cert.SerialNumber.Text(16))
	return nil
}

// MustSetCertificate is like SetCertificate but panics if operation fails.
func (client *Client) MustSetCertificate(certPEM, keyPem string) {
	if err := client.SetCertificate(certPEM, keyPem); err != nil {
		panic(err)
	}
}

type requestXml interface{}

type requestable interface {
	toXml(client *Client) requestXml
}

type requestJson interface{}

type jsonRequestable interface {
	toJson(client *Client) requestJson
}

type responsible interface {
	Success() bool
	AsError() error
}

func (client *Client) postJson(ctx context.Context, url string, object jsonRequestable, res responsible) error {
	jsonData, err := json.MarshalIndent(object.toJson(client), "", "  ")
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", applicationJson)
	req.Header.Set("Content-Type", applicationJson)
	auth, err := client.generateAuthorization(http.MethodPost, url, string(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)
	b, statusCode, err := client.post(req, jsonData)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, res)
	if err != nil {
		return err
	}
	if res.Success() {
		if statusCode != 200 {
			return JsonResponseError{
				Code:    "UNKNOWN",
				Message: "未知错误，状态：" + strconv.Itoa(statusCode),
			}
		}
		return nil
	} else {
		return res.AsError()
	}
}

func (client *Client) postXml(ctx context.Context, url string, object requestable, res responsible) error {
	xmlData, err := xml.MarshalIndent(object.toXml(client), "", "  ")
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(xmlData))
	if err != nil {
		return err
	}
	b, _, err := client.post(req, xmlData)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(b, res)
	if err != nil {
		return err
	}
	if res.Success() {
		return nil
	} else {
		return res.AsError()
	}
}

func (client *Client) post(httpReq *http.Request, reqData []byte) ([]byte, int, error) {
	if client.Debug {
		dump, err := httputil.DumpRequestOut(httpReq, true)
		if err != nil {
			return nil, 0, err
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
		return nil, 0, err
	}
	defer resp.Body.Close()
	if client.Debug {
		contentType := resp.Header.Get("Content-Type")
		dumpBody := strings.Contains(contentType, applicationJson) || strings.Contains(contentType, "text/")
		dump, err := httputil.DumpResponse(resp, dumpBody)
		if err != nil {
			return nil, resp.StatusCode, err
		}
		log.Println(string(dump))
	}
	b, err := ioutil.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}

func (client Client) generateSign(object interface{}) string {
	str, signType := generateStringToSign(object, client.Key)
	if client.Debug {
		log.Println("sign type", signType)
		log.Println("string to sign", str)
	}
	if signType == "HMAC-SHA256" {
		h := hmac.New(sha256.New, []byte(client.Key))
		h.Write([]byte(str))
		return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	}
	h := md5.New()
	h.Write([]byte(str))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func (client Client) generateAuthorization(method, url, reqBody string) (string, error) {
	nonce := randomStr(32)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	var signStr strings.Builder
	signStr.WriteString(method)
	signStr.WriteString("\n")
	signStr.WriteString(url[len(prefix):])
	signStr.WriteString("\n")
	signStr.WriteString(timestamp)
	signStr.WriteString("\n")
	signStr.WriteString(nonce)
	signStr.WriteString("\n")
	signStr.WriteString(reqBody)
	signStr.WriteString("\n")

	sign, err := client.sha256rsa2048sign([]byte(signStr.String()))
	if err != nil {
		return "", err
	}

	var authStr strings.Builder
	authStr.WriteString("WECHATPAY2-SHA256-RSA2048 ")
	authStr.WriteString(`mchid="`)
	authStr.WriteString(client.MchId)
	authStr.WriteString(`",`)
	authStr.WriteString(`nonce_str="`)
	authStr.WriteString(nonce)
	authStr.WriteString(`",`)
	authStr.WriteString(`signature="`)
	authStr.WriteString(base64.StdEncoding.EncodeToString(sign))
	authStr.WriteString(`",`)
	authStr.WriteString(`timestamp="`)
	authStr.WriteString(timestamp)
	authStr.WriteString(`",`)
	authStr.WriteString(`serial_no="`)
	authStr.WriteString(client.certSerialNumber)
	authStr.WriteString(`"`)

	return authStr.String(), nil
}

func (client Client) sha256rsa2048sign(data []byte) ([]byte, error) {
	h := crypto.SHA256.New()
	h.Write(data)
	pk := client.TLSClientConfig.Certificates[0].PrivateKey.(*rsa.PrivateKey)
	return rsa.SignPKCS1v15(cryptoRand.Reader, pk, crypto.SHA256, h.Sum(nil))
}

type Response struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
	NonceStr   string `xml:"nonce_str,omitempty"`
	Sign       string `xml:"sign,omitempty"`
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
		if r.ReturnCode != "" && r.ReturnMsg != "" {
			return r.ReturnCode + " (" + r.ReturnMsg + ")"
		}
		return "UNKNOWN"
	}
	return r.ErrCode + ": " + r.ErrCodeDes
}

type JsonResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (r JsonResponse) Success() bool {
	return r.Code == ""
}

type JsonResponseError JsonResponse

func (r JsonResponseError) Error() string {
	return r.Code + ": " + r.Message
}

func copyFields(from, to interface{}) {
	fromRV := reflect.ValueOf(from)
	fromRT := reflect.TypeOf(from)
	toRV := reflect.ValueOf(to).Elem()
	for i := 0; i < fromRT.NumField(); i++ {
		f := fromRT.Field(i)
		toRV.FieldByName(f.Name).Set(fromRV.FieldByName(f.Name))
	}
}

func generateStringToSign(s interface{}, key string) (stringToSign, signType string) {
	rv := reflect.ValueOf(s)
	rt := reflect.TypeOf(s)
	names := []string{}
	values := map[string]string{}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Name == "XMLName" {
			continue
		}
		name := f.Tag.Get("xml")
		if name == "-" {
			continue
		}
		if i := strings.Index(name, " "); i >= 0 {
			name = name[i+1:]
		}
		var isOmitempty bool
		if tokens := strings.Split(name, ","); len(tokens) > 1 {
			name = tokens[0]
			for _, t := range tokens[1:] {
				if t == "omitempty" {
					isOmitempty = true
					break
				}
			}
		}
		if name == "sign" {
			continue
		}
		if isOmitempty && rv.Field(i).IsZero() {
			continue
		}
		names = append(names, name)
		values[name] = fmt.Sprint(rv.Field(i).Interface())
		if name == "sign_type" {
			signType = values[name]
		}
	}
	sort.Strings(names)
	var buf bytes.Buffer
	for _, name := range names {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(name)
		buf.WriteByte('=')
		buf.WriteString(values[name])
	}
	buf.WriteString("&key=")
	buf.WriteString(key)
	stringToSign = buf.String()
	return
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

type JSAPIPayParams struct {
	AppId     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// Generate pay params for JSAPI.
func (client *Client) JSAPIPayParams(appId, prepayId string) *JSAPIPayParams {
	p := &JSAPIPayParams{
		AppId:     appId,
		TimeStamp: strconv.FormatInt(time.Now().Unix(), 10),
		NonceStr:  randomStr(32),
		Package:   "prepay_id=" + prepayId,
		SignType:  "MD5",
	}
	h := md5.New()
	str := "appId=" + p.AppId + "&nonceStr=" + p.NonceStr +
		"&package=" + p.Package + "&signType=" + p.SignType +
		"&timeStamp=" + p.TimeStamp + "&key=" + client.Key
	h.Write([]byte(str))
	p.PaySign = strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	return p
}

type Utc8Time time.Time

func (tm Utc8Time) String() string {
	return time.Time(tm).String()
}

func (tm Utc8Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(tm))
}

func (tm *Utc8Time) UnmarshalJSON(data []byte) error {
	var t time.Time
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}
	*tm = Utc8Time(t)
	return nil
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
