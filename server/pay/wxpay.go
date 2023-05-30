package pay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"time"
)

var WXClient *wechat.ClientV3

func InitWx() {
	mchid := viper.GetString("wxpay.mchid")
	serialNo := viper.GetString("wxpay.serialNo")
	privateKey := viper.GetString("wxpay.privateKey")
	apiV3Key := viper.GetString("wxpay.apiV3Key")

	b, err := os.ReadFile(privateKey)
	if err != nil {
		panic(err)
	}
	client, err := wechat.NewClientV3(mchid, serialNo, apiV3Key, string(b))
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
	notifyURL := viper.GetString("wxpay.notifyUrl")
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return "", err
	}
	expire := time.Now().Add(10 * time.Minute).Format(time.RFC3339)

	// 初始化 BodyMap
	bm := make(gopay.BodyMap)
	bm.Set("appid", viper.GetString("wxpay.appid")).
		Set("description", subject).
		Set("out_trade_no", orderID).
		Set("time_expire", expire).
		Set("notify_url", notifyURL).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", amt.Mul(decimal.NewFromInt(100)).BigInt().Uint64()).
				Set("currency", "CNY")
		})

	wxRsp, err := WXClient.V3TransactionNative(context.Background(), bm)
	if err != nil {
		return "", err
	}
	if wxRsp.Code != http.StatusOK {
		return "", fmt.Errorf(wxRsp.Error)
	}

	return wxRsp.Response.CodeUrl, nil
}
