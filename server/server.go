package server

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/redesblock/dataserver/dataservice"
	"github.com/redesblock/dataserver/docs"
	"github.com/redesblock/dataserver/server/routers"
	log "github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
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
	v1.POST("/buckets", routers.AddBucketHandler(db))

	v1.GET("/buckets/:id/objects", routers.GetBucketObjectsHandler(db))
	v1.GET("/buckets/:id/objects/:fid", routers.GetBucketObjectHandler(db))
	v1.DELETE("/buckets/:id/objects/:fid", routers.DeleteBucketObjectHandler(db))
	v1.POST("/buckets/:id/objects/:name", routers.AddBucketObjectHandler(db))

	v1.GET("/contract", routers.GetContractHandler(db))
	v1.GET("/buy/storage", routers.BuyStorageHandler(db))
	v1.GET("/buy/traffic", routers.BuyTrafficHandler(db))
	v1.GET("/bills/storage", routers.GetBillsStorageHandler(db))
	v1.POST("/bills/storage", routers.AddBillsStorageHandler(db))
	v1.GET("/bills/traffic", routers.GetBillsTrafficHandler(db))
	v1.POST("/bills/traffic", routers.AddBillsTrafficHandler(db))

	v1.GET("/asset/:id", routers.GetAssetHandler(db))
	v1.GET("/upload/:asset_id", routers.GetFileUploadHandler(db))
	v1.POST("/upload/:asset_id", routers.FileUploadHandler(db))
	v1.GET("/download/:cid", routers.GetFileDownloadHandler(db, node))

	uploadChan := make(chan string, 10)
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

	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(1).Day().At("00:00;12:00").Do(func() {

	})
	scheduler.Every(10).Minute().Do(func() {
		uploadChan <- "scheduler"
	})
	scheduler.StartAsync()

	//v1.POST("/upload", routers.UploadHandler(db))
	log.Info("starting server at port ", port)
	log.Fatal("starting server error: ", router.Run(port))
}
