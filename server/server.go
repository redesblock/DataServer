package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/redesblock/dataserver/models"
	"github.com/redesblock/dataserver/server/dispatcher"
	"github.com/redesblock/dataserver/server/pay"
	"github.com/redesblock/dataserver/server/routers"
	v1 "github.com/redesblock/dataserver/server/routers/api/v1"
	"github.com/spf13/viper"
	"io"
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
					ret, err := txStatus(item.Hash)
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
							if err := tx.Model(&models.Order{}).Where("hash = ?", item.Hash).Update("status", models.OrderSuccess).Error; err != nil {
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

func txStatus(hash string) (int, error) {
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionReceipt",
		"params":  []string{hash},
		"id":      1,
	}
	body, _ := json.Marshal(request)
	resp, err := http.Post(viper.GetString("bsc.rpc"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(resp.Status)
	}
	defer resp.Body.Close()

	bts, _ := io.ReadAll(resp.Body)
	jsonParsed, err := gabs.ParseJSON(bts)
	if err != nil {
		return 0, err
	}

	if jsonParsed.Exists("result", "status") {
		if blkHash := jsonParsed.Path("result.blockHash").Data().(string); len(blkHash) > 0 {
			if status := jsonParsed.Path("result.status").Data().(string); status == "0x1" {
				return 1, nil
			} else {
				return 2, nil
			}
		}
	}
	return 0, fmt.Errorf(jsonParsed.String())
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
	if err := job.db.Model(&models.Node{}).Order("area desc").Where("voucher_id <> ''").Where("usable = true").Find(&vouchers).Error; err != nil {
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
