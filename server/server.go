package server

import (
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/redesblock/dataserver/dataservice"
	"github.com/redesblock/dataserver/docs"
	"github.com/redesblock/dataserver/server/routers"
	log "github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func Start(port string, db *dataservice.DataService) {
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
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := router.Group("/api/v1")
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
	v1.POST("/buckets/:id/objects/:name", routers.AddBucketObjectHandler(db, node))

	uploadBill := make(chan string, 1024)
	v1.GET("/contract", routers.GetContractHandler(db))
	v1.GET("/buy/storage", routers.BuyStorageHandler(db))
	v1.GET("/buy/traffic", routers.BuyTrafficHandler(db))
	v1.GET("/bills/storage", routers.GetBillsStorageHandler(db))
	v1.POST("/bills/storage", routers.AddBillsStorageHandler(db, uploadBill))
	v1.GET("/bills/traffic", routers.GetBillsTrafficHandler(db))
	v1.POST("/bills/traffic", routers.AddBillsTrafficHandler(db, uploadBill))

	v1.GET("/asset/:id", routers.GetAssetHandler(db))
	v1.GET("/upload/:asset_id", routers.GetFileUploadHandler(db))
	v1.POST("/upload/:asset_id", routers.FileUploadHandler(db))
	v1.GET("/download/:cid", routers.GetFileDownloadHandler(db, node))

	uploadChan := make(chan string, 1024)
	v1.POST("/finish/:asset_id", routers.FinishFileUploadHandler(db, uploadChan))

	go func() {
		for {
			select {
			case <-uploadChan:
				var items []*dataservice.BucketObject
				db.Find(&dataservice.BucketObject{}).Where("size > 0").Where("asset_id != ''").Where("c_id = ''").Where("status = ?", dataservice.STATUS_UPLOADED).Find(&items)
				for _, item := range items {
					t := time.Now()
					cid, err := uploadFiles(node(), voucher(), item.AssetID, item.Name)
					if err != nil {
						log.Errorf("upload %s error %v", item.AssetID, err)
						continue
					}
					log.Infof("upload file %s, size %d, cid %s, elapse %v", item.AssetID, item.Size, cid, time.Now().Sub(t))
					item.Status = dataservice.STATUS_PINED
					item.CID = cid
					db.Save(&item)
				}
			}
		}
	}()

	go func() {
		txStatusFunc := func(hash string) int {
			retBody := strings.NewReader(fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionReceipt\",\"params\":[\"%s\"],\"id\":1}", hash))
			resp, err := http.Post("http://202.83.246.155:8575", "application/json", retBody)
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
		for {
			select {
			case <-uploadBill:
				var items []*dataservice.BillStorage
				db.Find(&dataservice.BillStorage{}).Where("status = ?", dataservice.TX_STATUS_PEND).Find(&items)
				for _, item := range items {
					item.Status = txStatusFunc(item.Hash)
					if err := db.Transaction(func(tx *gorm.DB) error {
						if item.Status == dataservice.TX_STATUS_SUCCESS {
							var user dataservice.User
							if ret := tx.Model(&dataservice.User{}).Where("id = ?", item.UserID).Find(&user); ret.Error != nil {
								return ret.Error
							} else if ret.RowsAffected == 0 {
								return fmt.Errorf("not found user")
							}
							user.TotalStorage += item.Size
							if err := tx.Save(&user).Error; err != nil {
								return err
							}
						}
						return tx.Save(&item).Error
					}); err != nil {
						log.Error("upload bill error ", err)
					}
				}

				var items2 []*dataservice.BillTraffic
				db.Find(&dataservice.BillTraffic{}).Where("status = ?", dataservice.TX_STATUS_PEND).Find(&items2)
				for _, item := range items2 {
					item.Status = txStatusFunc(item.Hash)
					if err := db.Transaction(func(tx *gorm.DB) error {
						if item.Status == dataservice.TX_STATUS_SUCCESS {
							var user dataservice.User
							if ret := tx.Model(&dataservice.User{}).Where("id = ?", item.UserID).Find(&user); ret.Error != nil {
								return ret.Error
							} else if ret.RowsAffected == 0 {
								return fmt.Errorf("not found user")
							}
							user.TotalTraffic += item.Size
							if err := tx.Save(&user).Error; err != nil {
								return err
							}
						}
						return tx.Save(&item).Error
					}); err != nil {
						log.Error("upload bill error ", err)
					}
				}
			}
		}
	}()

	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(1).Day().At("00:00;12:00").Do(func() {

	})
	scheduler.Every(10).Minute().Do(func() {
		uploadChan <- "scheduler"
	})
	scheduler.Every(15).Second().Do(func() {
		uploadBill <- "bill"
	})
	scheduler.StartAsync()

	//v1.POST("/upload", routers.UploadHandler(db))
	log.Info("starting server at port ", port)
	log.Fatal("starting server error: ", router.Run(port))
}
