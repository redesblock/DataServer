package pay

import (
	"context"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

var WXClient *wechat.ClientV3

func InitWx() {
	mchid := viper.GetString("wxpay.mchid")
	serialNo := viper.GetString("wxpay.serialNo")
	privateKey := viper.GetString("wxpay.privateKey")
	apiV3Key := viper.GetString("wxpay.apiV3Key")

	client, err := wechat.NewClientV3(mchid, serialNo, apiV3Key, privateKey)
	if err != nil {
		panic(err)
	}
	err = client.AutoVerifySign()
	if err != nil {
		panic(err)
	}

	WXClient = client
}

func WXTrade(subject, orderID, amount string) (string, error) {
	expire := time.Now().Add(10 * time.Minute).Format(time.RFC3339)
	notifyURL := viper.GetString("wx.notifyUrl")
	// 初始化 BodyMap
	bm := make(gopay.BodyMap)
	bm.Set("sp_appid", "sp_appid").
		Set("sp_mchid", "sp_mchid").
		Set("sub_mchid", "sub_mchid").
		Set("description", subject).
		Set("out_trade_no", orderID).
		Set("time_expire", expire).
		Set("notify_url", notifyURL).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", amount).
				Set("currency", "CNY")
		}).
		SetBodyMap("payer", func(bm gopay.BodyMap) {
			bm.Set("sp_openid", "asdas")
		})

	wxRsp, err := WXClient.V3TransactionJsapi(context.Background(), bm)
	if err != nil {
		return "", err
	}
	if wxRsp.Code != http.StatusOK {
		return "", err
	}
	//jsapi, err := WXClient.PaySignOfJSAPI("appid", "prepayid")

	return wxRsp.Response.PrepayId, nil
}
