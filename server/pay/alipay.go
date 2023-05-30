package pay

import (
	"github.com/redesblock/dataserver/models"
	"github.com/smartwalle/alipay/v3"
	"github.com/spf13/viper"
	"time"
)

var AlipayClient *alipay.Client

func InitAlipay() {
	appid := viper.GetString("alipay.appid")
	isprod := viper.GetBool("alipay.isprod")
	appkey := viper.GetString("alipay.app.privateKey")
	apppub := viper.GetString("alipay.app.publicKey")
	aliroot := viper.GetString("alipay.root")
	alipub := viper.GetString("alipay.publicKey")

	client, err := alipay.New(appid, appkey, isprod)
	if err != nil {
		panic(err)
	}
	client.LoadAppPublicCertFromFile(apppub)
	client.LoadAliPayRootCertFromFile(aliroot)
	client.LoadAliPayPublicCertFromFile(alipub)

	AlipayClient = client
}

func AliPayTrade(subject, orderID, amount string) (string, error) {
	var p = alipay.TradePagePay{}
	p.ReturnURL = viper.GetString("alipay.returnUrl")
	p.NotifyURL = viper.GetString("alipay.notifyUrl")

	p.Subject = subject
	p.OutTradeNo = orderID
	p.TotalAmount = amount
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	p.TimeExpire = time.Now().Add(10 * time.Minute).Format(models.TIME_FORMAT)
	url, err := AlipayClient.TradePagePay(p)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
