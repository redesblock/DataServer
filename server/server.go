package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/redesblock/dataserver/server/dispatcher"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"github.com/redesblock/dataserver/docs"
	"github.com/redesblock/dataserver/server/routers"
	log "github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func Start(port string, db *dataservice.DataService) {
	dispatcher.NewDispatcher(100).Run()

	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(gzip.Gzip(gzip.BestSpeed))
	router.Use(func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	})
	//router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := router.Group("/api/v1")
	v1.POST("/traffic", routers.AddReportTrafficHandler(db))
	v1.GET("/traffics", routers.GetReportTrafficsHandler(db))
	v1.GET("/traffics/{date}", routers.GetReportTrafficHandler(db))

	// v1.POST("/receipt", routers.AddReportReceiptHandler(db))

	v1.POST("/login", routers.LoginHandler(db))
	v1.Use(routers.JWTAuthMiddleware())
	v1.GET("/user", routers.GetUserHandler(db))
	v1.POST("/user", routers.AddUserHandler(db))
	v1.GET("/user/actions", routers.UserActionsHandler(db))

	v1.GET("/overview", routers.OverViewHandler(db))
	v1.GET("/daily/storage", routers.DailyStorageHandler(db))
	v1.GET("/daily/traffic", routers.DailyTrafficHandler(db))

	v1.GET("/networks", routers.GetNetWorksHandler(db))
	v1.GET("/areas", routers.GetAreasHandler(db))

	v1.GET("/buckets", routers.GetBucketsHandler(db))
	v1.GET("/buckets/:id", routers.GetBucketHandler(db))
	v1.DELETE("/buckets/:id", routers.DeleteBucketHandler(db))
	v1.POST("/buckets/:id", routers.UpdateBucketHandler(db))
	v1.POST("/buckets", routers.AddBucketHandler(db))

	v1.GET("/buckets/:id/objects", routers.GetBucketObjectsHandler(db))
	v1.GET("/buckets/:id/objects/:fid", routers.GetBucketObjectHandler(db))
	v1.DELETE("/buckets/:id/objects/:fid", routers.DeleteBucketObjectHandler(db))
	v1.POST("/buckets/:id/objects/:name", routers.AddBucketObjectHandler(db))

	uploadedTx := make(chan []string)
	v1.GET("/contract", routers.GetContractHandler(db))
	v1.GET("/buy/storage", routers.BuyStorageHandler(db))
	v1.GET("/buy/traffic", routers.BuyTrafficHandler(db))
	v1.GET("/bills/storage", routers.GetBillsStorageHandler(db))
	v1.POST("/bills/storage", routers.AddBillsStorageHandler(db, uploadedTx))
	v1.GET("/bills/traffic", routers.GetBillsTrafficHandler(db))
	v1.POST("/bills/traffic", routers.AddBillsTrafficHandler(db, uploadedTx))

	v1.GET("/asset/:id", routers.GetAssetHandler(db))
	v1.GET("/upload/:asset_id", routers.GetFileUploadHandler(db))
	v1.POST("/upload/:asset_id", routers.FileUploadHandler(db))
	v1.GET("/download/:cid", routers.GetFileDownloadHandler(db))

	uploadedAsset := make(chan []string, 512)
	v1.POST("/finish/:asset_id", routers.FinishFileUploadHandler(db, uploadedAsset))

	// update tx status
	go func() {
		txStatusFunc := func(hash string) int {
			retBody := strings.NewReader(fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionReceipt\",\"params\":[\"%s\"],\"id\":1}", hash))
			resp, err := http.Post(viper.GetString("bsc.rpc"), "application/json", retBody)
			if err != nil {
				log.Error("sync tx status error ", err)
				return dataservice.TX_STATUS_PEND
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Error("sync tx status error ", resp.Status)
				return dataservice.TX_STATUS_PEND
			}

			bts, _ := ioutil.ReadAll(resp.Body)
			jsonParsed, err := gabs.ParseJSON(bts)
			if err != nil {
				log.Error("sync tx status error ", resp.Status)
				return dataservice.TX_STATUS_PEND
			}

			if jsonParsed.Exists("result", "status") {
				if blkHash := jsonParsed.Path("result.blockHash").Data().(string); len(blkHash) > 0 {
					if status := jsonParsed.Path("result.status").Data().(string); status == "0x1" {
						return dataservice.TX_STATUS_SUCCESS
					} else {
						return dataservice.TX_STATUS_FAIL
					}
				}
			} else {
				log.Warn("sync tx status ", jsonParsed.String())
			}
			return dataservice.TX_STATUS_PEND
		}

		duration := time.Minute
		timer := time.NewTimer(duration)
		for {
			select {
			case <-timer.C:
				var items []*dataservice.BillStorage
				db.Find(&dataservice.BillStorage{}).Where("status = ?", dataservice.TX_STATUS_PEND).Find(&items)

				var items2 []*dataservice.BillTraffic
				db.Find(&dataservice.BillTraffic{}).Where("status = ?", dataservice.TX_STATUS_PEND).Find(&items2)

				var hashes []string
				for _, item := range items {
					hashes = append(hashes, item.Hash)
				}
				for _, item := range items2 {
					hashes = append(hashes, item.Hash)
				}
				if len(hashes) > 0 {
					select {
					case uploadedTx <- hashes:
					default:
					}
				}
				timer.Reset(duration)
			case hashes := <-uploadedTx:
				for _, hash := range hashes {
					var t1 dataservice.BillStorage
					var t2 dataservice.BillTraffic
					if ret := db.Find(&t1, "hash = ?", hash); ret.Error != nil {
						log.Error("upload tx status error ", ret.Error)
					} else if ret.RowsAffected > 0 {
						status := dataservice.TX_STATUS_PEND
						if t1.Status == dataservice.TX_STATUS_PEND {
							status = txStatusFunc(hash)
						}
						if status == dataservice.TX_STATUS_PEND {
							continue
						}
						t1.Status = status
						if err := db.Transaction(func(tx *gorm.DB) error {
							if t1.Status == dataservice.TX_STATUS_SUCCESS {
								var user dataservice.User
								if ret := tx.Model(&dataservice.User{}).Where("id = ?", t1.UserID).Find(&user); ret.Error != nil {
									return ret.Error
								} else if ret.RowsAffected == 0 {
									return fmt.Errorf("not found user")
								}
								user.TotalStorage += t1.Size
								if err := tx.Save(&user).Error; err != nil {
									return err
								}
							}
							return tx.Save(&t1).Error
						}); err != nil {
							log.Error("upload tx status error ", err)
						}
					} else if ret := db.Find(&t2, "hash = ?", hash); ret.Error != nil {
						log.Error("upload tx status error ", ret.Error)
					} else if ret.RowsAffected > 0 {
						status := dataservice.TX_STATUS_PEND
						if t2.Status == dataservice.TX_STATUS_PEND {
							status = txStatusFunc(hash)
						}
						if status == dataservice.TX_STATUS_PEND {
							continue
						}
						if err := db.Transaction(func(tx *gorm.DB) error {
							if t2.Status == dataservice.TX_STATUS_SUCCESS {
								var user dataservice.User
								if ret := tx.Model(&dataservice.User{}).Where("id = ?", t2.UserID).Find(&user); ret.Error != nil {
									return ret.Error
								} else if ret.RowsAffected == 0 {
									return fmt.Errorf("not found user")
								}
								user.TotalTraffic += t2.Size
								if err := tx.Save(&user).Error; err != nil {
									return err
								}
							}
							return tx.Save(&t2).Error
						}); err != nil {
							log.Error("upload tx status error ", err)
						}
					}
				}
			}
		}
	}()
	go func() {
		duration := 10 * time.Minute
		timer := time.NewTicker(duration)
		for {
			select {
			case <-timer.C:
				var vouchers []*dataservice.Voucher
				if err := db.Model(&dataservice.Voucher{}).Where("usable = true").Find(&vouchers).Error; err != nil {
					log.WithField("error", err).Errorf("load vouchers")
				}
				for _, voucher := range vouchers {
					usable, err := voucherUsable(voucher.Node, voucher.Voucher)
					if err != nil {
						log.WithField("error", err).Errorf("find voucher usable")
						continue
					}
					if voucher.Usable != usable {
						voucher.Usable = usable
						if err := db.Save(&voucher).Error; err != nil {
							log.WithField("error", err).Errorf("save voucher")
						}
					}
				}

				var assets []string
				db.Transaction(func(tx *gorm.DB) error {
					var items []*dataservice.BucketObject
					tx.Find(&dataservice.BucketObject{}).Where("size > 0").Where("c_id = ''").Where("status = ?", dataservice.STATUS_UPLOADED).Find(&items)
					if len(items) == 0 {
						return nil
					}
					for _, item := range items {
						item.UplinkProgress = 10
						assets = append(assets, item.AssetID)
					}
					return tx.Save(&items).Error
				})

				if len(assets) > 0 {
					select {
					case uploadedAsset <- assets:
					default:
					}
				}
				timer.Reset(duration)
			case assets := <-uploadedAsset:
				for _, asset := range assets {
					dispatcher.JobQueue <- &AsssetJob{
						asset: asset,
						db:    db,
					}
				}
			}
		}
	}()

	log.Info("starting server at port ", port)
	log.Fatal("starting server error: ", router.Run(port))
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

	bts, _ := ioutil.ReadAll(response.Body)
	var ret map[string]interface{}
	if err := json.Unmarshal(bts, &ret); err != nil {
		return false, err
	}
	return ret["usable"].(bool), nil
}

type AsssetJob struct {
	asset string
	db    *dataservice.DataService
}

func (job *AsssetJob) Do() error {
	var item *dataservice.BucketObject
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

	var vouchers []*dataservice.Voucher
	if err := job.db.Model(&dataservice.Voucher{}).Order("area desc").Where("usable = true").Find(&vouchers).Error; err != nil {
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
		cid, err := uploadFiles(voucher.Node, voucher.Voucher, item.AssetID, item.Name)
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
		item.Status = dataservice.STATUS_PIN
		item.UplinkProgress = 100
		item.CID = hash
		return tx.Save(item).Error
	}); err != nil {
		log.Errorf("upload file %s error %v", job.asset, err)
		return err
	}
	return nil
}
