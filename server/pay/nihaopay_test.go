package pay

import (
	"fmt"
	"github.com/spf13/viper"
	"testing"
)

func TestNihaoPay(t *testing.T) {
	viper.Set("nihaopay.isProd", false)
	viper.Set("nihaopay.returnUrl", "https://mopdstor.com/#/billing/index")
	viper.Set("nihaopay.notifyUrl", "https://mopdstor.com/api/v1/wxpay/notify")
	viper.Set("nihaopay.key", "f432bd12b52339667f24bcac1ad323a8b378755fd4ec48b75b5c4a4b08df1b76")
	viper.Set("nihaopay.currency", "USD")
	viper.Set("nihaopay.terminal", "WAP")
	fmt.Println(NihaoPayTrade("sss", "jkh25jh1348fd89sg", "cny", "100", "unionpay", ""))
}
