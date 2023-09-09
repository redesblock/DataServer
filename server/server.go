package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/redesblock/dataserver/models"
	"github.com/redesblock/dataserver/server/dispatcher"
	"github.com/redesblock/dataserver/server/pay"
	"github.com/redesblock/dataserver/server/routers"
	v1 "github.com/redesblock/dataserver/server/routers/api/v1"
	"github.com/shopspring/decimal"
	"github.com/smartwalle/alipay/v3"
	"github.com/spf13/viper"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Start(port string, db *gorm.DB) {
	v1.MAIL_HOST = viper.GetString("email.host")
	v1.MAIL_PORT = viper.GetInt("email.port")
	v1.MAIL_USER = viper.GetString("email.user")
	v1.MAIL_PWD = viper.GetString("email.pwd")

	dispatcher.NewDispatcher(100).Run()
	gin.SetMode(gin.DebugMode)
	pay.Init()
	r := routers.InitRouter(db)
	// update tx status
	go func() {
		duration := time.Second * 10
		timer := time.NewTimer(duration)
		for {
			select {
			case <-timer.C:
				if err := db.Model(&models.UserCoupon{}).Where("status = ?", models.UserCouponStatus_Normal).Where("end_time != ? AND end_time < ?", models.UnlimitedTime, time.Now()).Update("status", models.UserCouponStatus_Expired).Error; err != nil {
					log.Errorf("sync user coupon status: %s", err)
				}
				if err := db.Model(&models.Coupon{}).Where("status = ?", models.CouponStatus_NotStart).Where("reserve > 0").Where("start_time != ? AND start_time <= ?", models.UnlimitedTime, time.Now()).Update("status", models.CouponStatus_InProcess).Error; err != nil {
					log.Errorf("sync coupon status: %s", err)
				}
				//if err := db.Model(&models.Coupon{}).Where("reserve = 0").Update("status", models.CouponStatus_Completed).Error; err != nil {
				//	log.Errorf("sync coupon status: %s", err)
				//	return
				//}
				if err := db.Model(&models.Coupon{}).Where("status != ?", models.CouponStatus_Expired).Where("reserve > 0").Where("end_time != ? AND end_time < ?", models.UnlimitedTime, time.Now()).Update("status", models.CouponStatus_Expired).Error; err != nil {
					log.Errorf("sync coupon status: %s", err)
				}

				var items []*models.Order
				if err := db.Find(&models.Order{}).Where("payment = ?", models.PaymentChannel_Crypto).Where("status = ?", models.OrderPending).Find(&items).Error; err != nil {
					log.Errorf("sync tx status: %s", err)
				}
				for _, item := range items {
					ret, from, to, act, err := txStatus(item.Hash)
					if err != nil {
						log.Errorf("sync tx status: %s", err)
						continue
					}
					if ret == 1 {
						if err := db.Transaction(func(tx *gorm.DB) error {
							var user models.User
							if ret := tx.Model(&models.User{}).Where("id = ?", item.UserID).Find(&user); ret.Error != nil {
								return ret.Error
							} else if ret.RowsAffected == 0 {
								return fmt.Errorf("not found user")
							}
							user.TotalStorage += item.Quantity
							if err := tx.Save(&user).Error; err != nil {
								return err
							}
							if item.UserCouponID > 0 {
								var userCoupon models.UserCoupon
								if ret := tx.Model(&models.UserCoupon{}).Where("id = ?", item.UserCouponID).Find(&userCoupon); ret.Error != nil {
									return ret.Error
								} else if ret.RowsAffected == 0 {
									return fmt.Errorf("not found user coupon")
								}
								userCoupon.Status = models.UserCouponStatus_Used
								if err := tx.Save(&userCoupon).Error; err != nil {
									return err
								}
							}

							if err := tx.Model(&models.Order{}).Where("hash = ?", item.Hash).Updates(map[string]interface{}{"payment_account": from, "receive_account": to, "status": models.OrderSuccess, "payment_amount": act}).Error; err != nil {
								return err
							}
							return nil
						}); err != nil {
							log.Errorf("sync tx status: %s", err)
						}

					}
					if ret == 2 {
						if err := db.Model(&models.Order{}).Where("hash = ?", item.Hash).Update("status", models.OrderFailed).Error; err != nil {
							log.Errorf("sync tx status: %s", err)
						}
					}
				}
				timer.Reset(duration)
			}
		}
	}()

	// update voucher status
	go func() {
		duration := time.Minute * 10
		timer := time.NewTimer(duration)
		for {
			select {
			case <-timer.C:
				var vouchers []*models.Node
				if err := db.Model(&models.Node{}).Where("voucher_id != ''").Where("usable = true").Find(&vouchers).Error; err != nil {
					log.Errorf("sync voucher status: %s", err)
				}
				for _, voucher := range vouchers {
					usable, err := voucherUsable(voucher.IP, voucher.VoucherID)
					if err != nil {
						log.Errorf("sync voucher status: %s", err)
						continue
					}
					if voucher.Usable != usable {
						voucher.Usable = usable
						if err := db.Model(&voucher).Where("voucher_id = ?", voucher.VoucherID).Update("usable = ?", usable).Error; err != nil {
							log.Errorf("sync voucher status: %s", err)
						}
					}
				}
				timer.Reset(duration)
			}
		}
	}()

	// upload file
	go func() {
		duration := time.Second * 10
		timer := time.NewTicker(duration)
		for {
			select {
			case <-timer.C:
				var items []*models.BucketObject
				db.Find(&models.BucketObject{}).Where("size > 0").Where("c_id = ''").Where("status = ?", models.STATUS_UPLOADED).Find(&items)
				for _, item := range items {
					item.UplinkProgress = 10
					db.Save(item)
					dispatcher.JobQueue <- &AsssetJob{
						asset: item.AssetID,
						db:    db,
					}
				}

				timer.Reset(duration)
			}
		}
	}()

	// upload order
	go func() {
		timer := time.NewTimer(10 * time.Minute)
		for {
			select {
			case <-timer.C:
				var items []*models.Order
				if err := db.Where("status != ?", models.OrderSuccess).Where("created_at between ? and ?", time.Now().Add(-time.Hour*24), time.Now()).Find(&items).Error; err != nil {
					log.Errorf("sync order status: %s", err)
				}
				for _, item := range items {
					switch item.Payment {
					case models.PaymentChannel_WeChat:
						resp, err := pay.WXClient.V3TransactionQueryOrder(context.Background(), wechat.OutTradeNo, item.OrderID)
						if err != nil {
							log.Errorf("sync order status: %s", err)
						}

						if resp.Response.TradeState == wechat.TradeStateSuccess {
							item.Status = models.OrderSuccess
							item.PaymentID = resp.Response.TransactionId
							item.PaymentAccount = resp.Response.Payer.Openid
							item.PaymentAmount = decimal.NewFromInt(int64(resp.Response.Amount.PayerTotal)).Div(decimal.NewFromInt(100)).String()
							item.PaymentTime, _ = time.Parse(models.TIME_FORMAT, resp.Response.SuccessTime)
							if err := db.Transaction(func(tx *gorm.DB) error {
								var user models.User
								if ret := tx.Model(&models.User{}).Where("id = ?", item.UserID).Find(&user); ret.Error != nil {
									return ret.Error
								} else if ret.RowsAffected == 0 {
									return fmt.Errorf("not found user")
								}
								user.TotalStorage += item.Quantity
								if err := tx.Save(&user).Error; err != nil {
									return err
								}
								if item.UserCouponID > 0 {
									var userCoupon models.UserCoupon
									if ret := tx.Model(&models.UserCoupon{}).Where("id = ?", item.UserCouponID).Find(&userCoupon); ret.Error != nil {
										return ret.Error
									} else if ret.RowsAffected == 0 {
										return fmt.Errorf("not found user coupon")
									}
									userCoupon.Status = models.UserCouponStatus_Used
									if err := tx.Save(&userCoupon).Error; err != nil {
										return err
									}
								}
								return tx.Save(&item).Error
							}); err != nil {
								log.Errorf("sync order status: %s", err)
							}
						}
					case models.PaymentChannel_Alipay:
						req := alipay.TradeQuery{}
						req.OutTradeNo = item.OrderID
						resp, err := pay.AlipayClient.TradeQuery(req)
						if err != nil {
							log.Errorf("sync order status: %s", err)
						}
						if resp.TradeStatus == alipay.TradeStatusSuccess {
							item.Status = models.OrderSuccess
							item.PaymentID = resp.TradeNo
							item.PaymentAccount = resp.BuyerLogonId
							item.ReceiveAccount = resp.StoreName
							item.PaymentAmount = resp.TotalAmount
							item.PaymentTime, _ = time.Parse(models.TIME_FORMAT, resp.SendPayDate)
							if err := db.Transaction(func(tx *gorm.DB) error {
								var user models.User
								if ret := tx.Model(&models.User{}).Where("id = ?", item.UserID).Find(&user); ret.Error != nil {
									return ret.Error
								} else if ret.RowsAffected == 0 {
									return fmt.Errorf("not found user")
								}
								user.TotalStorage += item.Quantity
								if err := tx.Save(&user).Error; err != nil {
									return err
								}
								if item.UserCouponID > 0 {
									var userCoupon models.UserCoupon
									if ret := tx.Model(&models.UserCoupon{}).Where("id = ?", item.UserCouponID).Find(&userCoupon); ret.Error != nil {
										return ret.Error
									} else if ret.RowsAffected == 0 {
										return fmt.Errorf("not found user coupon")
									}
									userCoupon.Status = models.UserCouponStatus_Used
									if err := tx.Save(&userCoupon).Error; err != nil {
										return err
									}
								}
								return tx.Save(&item).Error
							}); err != nil {
								log.Errorf("sync order status: %s", err)
							}
						}
					case models.PaymentChannel_Stripe:
					case models.PaymentChannel_NihaoPay_UnionPay, models.PaymentChannel_NihaoPay_WeChat, models.PaymentChannel_NihaoPay_Alipay:
						ret, err := pay.NihaoPayQuery(item.OrderID)
						if err != nil {
							log.Errorf("sync order status: %s", err)
						}
						updated := false
						status := ret["status"]
						switch status {
						case "success":
							item.PaymentID = ret["id"].(string)
							amount, ok := ret["amount"].(float64)
							if !ok {
								amount, _ = ret["rmb_amount"].(float64)
							}
							item.PaymentAccount = decimal.NewFromFloat(amount).Div(decimal.NewFromInt(100)).String()
							item.PaymentTime, _ = time.Parse(time.RFC3339, ret["time"].(string))
							item.Status = models.OrderSuccess
							updated = true
						case "failure":
							if item.Status != models.OrderFailed {
								item.Status = models.OrderFailed
								updated = true
							}
						case "pending":
							if item.Status != models.OrderPending {
								item.Status = models.OrderPending
								updated = true
							}
						}
						if updated {
							if err := db.Transaction(func(tx *gorm.DB) error {
								var user models.User
								if ret := tx.Model(&models.User{}).Where("id = ?", item.UserID).Find(&user); ret.Error != nil {
									return ret.Error
								} else if ret.RowsAffected == 0 {
									return fmt.Errorf("not found user")
								}
								user.TotalStorage += item.Quantity
								if err := tx.Save(&user).Error; err != nil {
									return err
								}
								if item.UserCouponID > 0 {
									var userCoupon models.UserCoupon
									if ret := tx.Model(&models.UserCoupon{}).Where("id = ?", item.UserCouponID).Find(&userCoupon); ret.Error != nil {
										return ret.Error
									} else if ret.RowsAffected == 0 {
										return fmt.Errorf("not found user coupon")
									}
									userCoupon.Status = models.UserCouponStatus_Used
									if err := tx.Save(&userCoupon).Error; err != nil {
										return err
									}
								}
								return tx.Save(&item).Error
							}); err != nil {
								log.Errorf("sync order status: %s", err)
							}
						}
					}
				}
			}
			timer.Reset(10 * time.Minute)
		}
	}()

	log.Info("starting server at port ", port)
	log.Fatal("starting server error: ", r.Run(port))
}

func voucherUsable(node string, voucher string) (bool, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	response, err := client.Get("http://" + node + ":1685" + "/stamps/" + voucher)
	if err != nil {
		return false, err
	}
	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf(response.Status)
	}
	defer response.Body.Close()

	bts, _ := io.ReadAll(response.Body)
	var ret map[string]interface{}
	if err := json.Unmarshal(bts, &ret); err != nil {
		return false, err
	}
	return ret["usable"].(bool), nil
}

func txStatus(hash string) (int, string, string, string, error) {
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionReceipt",
		"params":  []string{hash},
		"id":      1,
	}
	body, _ := json.Marshal(request)
	resp, err := http.Post(viper.GetString("bsc.rpc"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, "", "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, "", "", "", fmt.Errorf(resp.Status)
	}
	defer resp.Body.Close()

	bts, _ := io.ReadAll(resp.Body)
	jsonParsed, err := gabs.ParseJSON(bts)
	if err != nil {
		return 0, "", "", "", err
	}

	if jsonParsed.Exists("result", "status") {
		from := jsonParsed.Path("result.from").Data().(string)
		to := jsonParsed.Path("result.to").Data().(string)
		token := jsonParsed.Path("result.logs.0.address").Data().(string)
		div := decimal.New(1, 18)
		switch token {
		case viper.GetString("price.mop"):
		case viper.GetString("price.usdt"):
			div = decimal.New(1, 8)
		}
		act := "0"
		num, ok := new(big.Int).SetString(jsonParsed.Path("result.logs.0.data").Data().(string)[2:], 16)
		if !ok {
			log.Error("logs", jsonParsed.Path("result.logs.0.data").String(), err)
		} else {
			act = decimal.NewFromBigInt(num, 0).Div(div).String()
		}
		if blkHash := jsonParsed.Path("result.blockHash").Data().(string); len(blkHash) > 0 {
			if status := jsonParsed.Path("result.status").Data().(string); status == "0x1" {
				return 1, from, to, act, nil
			} else {
				return 2, from, to, act, nil
			}
		}
	}
	return 0, "", "", "", fmt.Errorf(jsonParsed.String())
}

type AsssetJob struct {
	asset string
	db    *gorm.DB
}

func (job *AsssetJob) Do() error {
	var item *models.BucketObject
	if err := job.db.Transaction(func(tx *gorm.DB) error {
		if ret := tx.Find(&item, "asset_id = ?", job.asset); ret.Error != nil {
			return ret.Error
		} else if ret.RowsAffected == 0 {
			return nil
		}
		item.UplinkProgress = 50
		return tx.Save(item).Error
	}); err != nil {
		log.Errorf("upload file %s error %v", job.asset, err)
		return err
	}

	var vouchers []*models.Node
	if err := job.db.Model(&models.Node{}).Order("zone desc").Where("voucher_id <> ''").Where("usable = true").Find(&vouchers).Error; err != nil {
		log.Errorf("upload file %s error %v", job.asset, err)
		return err
	}
	if len(vouchers) == 0 {
		err := fmt.Errorf("no usable vouchers")
		log.Errorf("upload file %s error %v", job.asset, err)
		return err
	}

	voucherCnt := len(vouchers)
	voucherIndex := 0
	hash := ""
	for i := 0; i < voucherCnt; i++ {
		voucherIndex += i
		voucher := vouchers[voucherIndex%voucherCnt]
		t := time.Now()
		cid, err := uploadFiles(voucher.IP, voucher.VoucherID, item.AssetID, item.Name)
		if err != nil {
			log.Errorf("upload file %s error %v", job.asset, err)
			return err
		}
		hash = cid
		log.Infof("upload file %s, size %d, cid %s, elapse %v", item.AssetID, item.Size, cid, time.Now().Sub(t))
		break
	}

	if err := job.db.Transaction(func(tx *gorm.DB) error {
		if ret := tx.Find(&item, "asset_id = ?", job.asset); ret.Error != nil {
			return ret.Error
		} else if ret.RowsAffected == 0 {
			return nil
		}
		item.Status = models.STATUS_PIN
		item.UplinkProgress = 100
		item.CID = hash
		return tx.Save(item).Error
	}); err != nil {
		log.Errorf("upload file %s error %v", job.asset, err)
		return err
	}
	return nil
}
