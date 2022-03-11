# wxpayslim

## Transfer

```go
import "github.com/caiguanhao/wxpayslim"

client = wxpayslim.NewClient("1111111111", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")

// client.Debug = true

client.SetCertificate("-----BEGIN CERTIFICATE-----\n...", "-----BEGIN PRIVATE KEY-----\n...")

ctx := context.Background()

resp, err := client.Transfer(ctx, wxpayslim.TransferRequest{
	AppId:          "wxxxxxxxxxxxxxxxxx",
	PartnerTradeNo: "TESTz20220311z111122",
	OpenId:         "oAxxxxxxxxxxxxxxxxxxxxxxxxxx",
	Amount:         100,
	Desc:           "one-yuan",
})

// resp = {
//   ReturnCode:SUCCESS
//   ReturnMsg:
//   NonceStr:ZneDMNUuaOidCoYaQ2DAAVOkWP4kOUyf
//   ResultCode:SUCCESS
//   ErrCode:
//   ErrCodeDes:
//   AppId:wxxxxxxxxxxxxxxxxx
//   MchId:1111111111
//   DeviceInfo:
//   PartnerTradeNo:TESTz20220311z111122
//   PaymentNo:10000000000000000000000000000000
//   PaymentTime:2022-03-11 11:11:23 +0800 UTC+8
// }

resp2, err := client.TransferQuery(ctx, wxpayslim.TransferQueryRequest{
	AppId:          "wxxxxxxxxxxxxxxxxx",
	PartnerTradeNo: "TESTz20220311z111122",
})

// resp2 = {
//   ReturnCode:SUCCESS
//   ReturnMsg:
//   NonceStr:
//   ResultCode:SUCCESS
//   ErrCode:
//   ErrCodeDes:
//   AppId:wxxxxxxxxxxxxxxxxx
//   MchId:1111111111
//   DetailId:10000000000000000000000000000000
//   Status:SUCCESS
//   Reason:
//   OpenId:oAxxxxxxxxxxxxxxxxxxxxxxxxxx
//   TransferName:
//   PartnerTradeNo:TESTz20220311z111122
//   PaymentAmount:100
//   TransferTime:2022-03-11 11:11:22 +0800 UTC+8
//   PaymentTime:2022-03-11 11:11:23 +0800 UTC+8
//   Desc:one-yuan
// }
```
