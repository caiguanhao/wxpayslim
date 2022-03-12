package wxpayslim

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
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
	"strings"
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

type requestXml interface{}

type requestable interface {
	toXml(client *Client) requestXml
}

type responsible interface {
	Success() bool
	AsError() error
}

func (client *Client) postXml(ctx context.Context, url string, object requestable, res responsible) error {
	xmlData, err := xml.MarshalIndent(object.toXml(client), "", "  ")
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(xmlData))
	if err != nil {
		return err
	}
	if client.Debug {
		dump, err := httputil.DumpRequestOut(httpReq, true)
		if err != nil {
			return err
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
		return err
	}
	defer resp.Body.Close()
	if client.Debug {
		dumpBody := strings.Contains(resp.Header.Get("Content-Type"), "text/")
		dump, err := httputil.DumpResponse(resp, dumpBody)
		if err != nil {
			return err
		}
		log.Println(string(dump))
	}
	b, err := ioutil.ReadAll(resp.Body)
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

func (client Client) generateSign(object interface{}) string {
	str := structToString(object) + "&key=" + client.Key
	return fmt.Sprintf("%X", md5.Sum([]byte(str)))
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

func copyFields(from, to interface{}) {
	fromRV := reflect.ValueOf(from)
	fromRT := reflect.TypeOf(from)
	toRV := reflect.ValueOf(to).Elem()
	for i := 0; i < fromRT.NumField(); i++ {
		f := fromRT.Field(i)
		toRV.FieldByName(f.Name).Set(fromRV.FieldByName(f.Name))
	}
}

func structToString(s interface{}) string {
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
		if tokens := strings.Split(name, ","); len(tokens) > 1 {
			name = tokens[0]
		}
		if name == "sign" {
			continue
		}
		names = append(names, name)
		values[name] = fmt.Sprint(rv.Field(i).Interface())
	}
	sort.Strings(names)
	var buf bytes.Buffer
	for _, name := range names {
		if values[name] == "" {
			continue
		}
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(name)
		buf.WriteByte('=')
		buf.WriteString(values[name])
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
