package pay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"strings"
)

func NihaoPayTrade(subject, orderID, currency, amount, vendor, note string) (string, error) {
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return "", err
	}
	amountStr := "amount"
	currency = strings.ToUpper(currency)
	if currency == "CNY" {
		amountStr = "rmb_amount"
	}
	apiUrl := "https://apitest.nihaopay.com/v1.2/transactions/securepay"
	if isProd := viper.GetBool("nihaopay.isProd"); isProd {
		apiUrl = "https://api.nihaopay.com/v1.2/transactions/securepay"
	}
	requestData := []byte(fmt.Sprintf(`{
		"%s": %d,
        "currency": "%s",
		"vendor": "%s",
		"ipn_url":"%s",
		"callback_url":"%s",
		"reference": "%s",
		"note": "%s",
		"description": "%s",
		"timeout": 10,
		"terminal": "%s"
    }`, amountStr, amt.Mul(decimal.NewFromInt(100)).BigInt().Int64(), viper.GetString("nihaopay.currency"), vendor, viper.GetString("nihaopay.notifyUrl"), viper.GetString("nihaopay.returnUrl"), orderID, note, subject, viper.GetString("nihaopay.terminal")))

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("nihaopay.key")))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bts, _ := io.ReadAll(resp.Body)

	// 处理响应
	if resp.StatusCode == http.StatusOK {
		// 处理成功支付的响应
	} else {
		fmt.Println("Payment failed. Status code:", resp.StatusCode)
		// 处理支付失败的响应
		return "", fmt.Errorf("%s %s", resp.Status, string(bts))
	}

	return string(bts), nil
}

func NihaoPayQuery(orderID string) (map[string]interface{}, error) {
	apiUrl := "https://apitest.nihaopay.com/v1.2/transactions/merchant/"
	if isProd := viper.GetBool("nihaopay.isProd"); isProd {
		apiUrl = "https://api.nihaopay.com/v1.2/transactions/merchant/"
	}

	req, err := http.NewRequest("GET", apiUrl+orderID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("nihaopay.key")))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bts, _ := io.ReadAll(resp.Body)

	// 处理响应
	if resp.StatusCode == http.StatusOK {
		// 处理成功支付的响应
	} else {
		fmt.Println("Payment Query failed. Status code:", resp.StatusCode)
		// 处理支付失败的响应
		return nil, fmt.Errorf("%s %s", resp.Status, string(bts))
	}

	var ret map[string]interface{}
	json.Unmarshal(bts, &ret)

	return ret, nil
}
