package pay

import (
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"github.com/spf13/viper"
	"io"
)

func AliPayClient() (*alipay.Client, error) {
	client, err := alipay.NewClient(viper.GetString("alipay.appid"), viper.GetString("alipay.privateKey"), viper.GetBool("alipay.isProd"))
	if err != nil {
		return nil, err
	}
	client.DebugSwitch = gopay.DebugOn
	client.SetLocation(alipay.LocationShanghai).
		SetCharset(alipay.UTF8).
		SetSignType(alipay.RSA2).
		SetReturnUrl(viper.GetString("alipay.returnUrl")).
		SetNotifyUrl(viper.GetString("alipay.notifyUrl"))

	client.AutoVerifySign([]byte(viper.GetString("alipay.PublicKeyContent")))
	err = client.SetCertSnByPath("appCertPublicKey.crt", "alipayRootCert.crt", "alipayCertPublicKey_RSA2.crt")
	if err != nil {
		return nil, err
	}
	return client, nil
}
