package routers

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	v1 "github.com/redesblock/dataserver/server/routers/api/v1"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	"net/http"
)

func InitRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.SetTrustedProxies(nil)
	r.Use(gzip.Gzip(gzip.BestSpeed))
	r.Use(func(c *gin.Context) {
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
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiv1 := r.Group("/api/v1")
	{
		apiv1.GET("/alipay", v1.AlipayTest(db))
		apiv1.GET("/wxpay", v1.WxpayTest(db))

		apiv1.POST("/traffic", v1.AddReportTraffic(db))
		apiv1.GET("/traffics", v1.GetReportTraffics(db))
		apiv1.POST("/login", v1.Login(db))
		apiv1.POST("/login2", v1.Login2(db))
		apiv1.POST("/forgot", v1.Forgot(db))
		apiv1.POST("/reset", v1.Reset(db))

		apiv1.POST("/alipay/notify", v1.AlipayNotify(db))
		apiv1.POST("/wxpay/notify", v1.AlipayNotify(db))

		apiv1.Use(v1.JWTAuthMiddleware())
		apiv1.GET("/user", v1.GetUserHandler(db))
		apiv1.GET("/user/signIn", v1.GetSignInSwitch(db))
		apiv1.GET("/user/signedIn", v1.GetSignedIn(db))
		apiv1.GET("/user/claimed", v1.GetClaimed(db))
		apiv1.GET("/user/unclaimed", v1.GetUnclaimed(db))
		apiv1.GET("/user/unclaimed/:id", v1.GetClaim(db))

		apiv1.GET("/products", v1.GetProducts(db))
		apiv1.GET("/products/:id", v1.GetProduct(db))

		apiv1.GET("/special_products", v1.GetSpecialProducts(db))
		apiv1.GET("/special_products/:id", v1.GetSpecialProduct(db))

		apiv1.GET("/currencies", v1.GetCurrencies(db))
		apiv1.GET("/currencies/:id", v1.GetCurrency(db))

		apiv1.POST("/user", v1.AddUserHandler(db))
		apiv1.GET("/user/actions", v1.UserActionsHandler(db))
		apiv1.GET("/overview", v1.OverViewHandler(db))
		apiv1.GET("/daily/storage", v1.DailyStorageHandler(db))
		apiv1.GET("/daily/traffic", v1.DailyTrafficHandler(db))
		apiv1.GET("/networks", v1.GetNetWorksHandler(db))
		apiv1.GET("/areas", v1.GetAreasHandler(db))

		apiv1.GET("/buckets", v1.GetBucketsHandler(db))
		apiv1.GET("/buckets/:id", v1.GetBucketHandler(db))
		apiv1.DELETE("/buckets/:id", v1.DeleteBucketHandler(db))
		apiv1.POST("/buckets/:id", v1.UpdateBucketHandler(db))
		apiv1.POST("/buckets", v1.AddBucketHandler(db))
		apiv1.GET("/buckets/:id/objects", v1.GetBucketObjectsHandler(db))
		apiv1.GET("/buckets/:id/objects/:fid", v1.GetBucketObjectHandler(db))
		apiv1.DELETE("/buckets/:id/objects/:fid", v1.DeleteBucketObjectHandler(db))
		apiv1.POST("/buckets/:id/objects/:name", v1.AddBucketObjectHandler(db))

		apiv1.GET("/asset/:id", v1.GetAssetHandler(db))
		apiv1.GET("/upload/:asset_id", v1.GetFileUploadHandler(db))
		apiv1.POST("/upload/:asset_id", v1.FileUploadHandler(db))
		apiv1.POST("/finish/:asset_id", v1.FinishFileUploadHandler(db))

		apiv1.GET("/download/:cid", v1.GetFileDownloadHandler(db))

		apiv1.GET("/contract", v1.GetERC20ContractHandler(db))
		apiv1.GET("/bills/storage", v1.GetBillsStorageHandler(db))
		apiv1.GET("/bills/traffic", v1.GetBillsTrafficHandler(db))
		apiv1.POST("/buy/storage", v1.BuyStorageHandler(db))
		apiv1.POST("/buy/traffic", v1.BuyTrafficHandler(db))
		apiv1.POST("/bills/storage", v1.AddBillsStorageHandler(db))
		apiv1.POST("/bills/traffic", v1.AddBillsTrafficHandler(db))
		apiv1.GET("/bills/traffic/:id", v1.GetOrder(db))
		apiv1.GET("/bills/storage/:id", v1.GetOrder(db))

		apiv1.Use(v1.JWTAuthMiddleware2())
		apiv1.PUT("/currencies/:id", v1.EditCurrency(db))
		apiv1.DELETE("/currencies/:id", v1.DeleteCurrency(db))

		apiv1.GET("/users", v1.GetUsers(db))
		apiv1.GET("/operators", v1.GetOperators(db))
		apiv1.GET("/users/:id", v1.GetUser(db))
		apiv1.POST("/users", v1.AddUser(db))
		apiv1.PUT("/users/:id", v1.EditUser(db))
		apiv1.DELETE("/users/:id", v1.DeleteUser(db))

		apiv1.GET("/orders", v1.GetOrders(db))
		apiv1.GET("/orders/:id", v1.GetOrder(db))

		apiv1.GET("/files", v1.GetFiles(db))

		apiv1.GET("/nodes", v1.GetNodes(db))
		apiv1.GET("/nodes/:id", v1.GetNode(db))
		apiv1.POST("/nodes", v1.AddNode(db))
		apiv1.PUT("/nodes/:id", v1.EditNode(db))
		apiv1.DELETE("/nodes/:id", v1.DeleteNode(db))
		apiv1.GET("/nodes/storage", v1.NodeStorageHandler(db))
		apiv1.GET("/nodes/traffic", v1.NodeTrafficHandler(db))

		apiv1.GET("/coupons", v1.GetCoupons(db))
		apiv1.GET("/coupons/:id", v1.GetCoupon(db))
		apiv1.POST("/coupons", v1.AddCoupon(db))
		apiv1.PUT("/coupons/:id", v1.EditCoupon(db))
		apiv1.DELETE("/coupons/:id", v1.DeleteCoupon(db))

		apiv1.POST("/special_products", v1.AddSpecialProduct(db))
		apiv1.PUT("/special_products/:id", v1.EditSpecialProduct(db))
		apiv1.DELETE("/special_products/:id", v1.DeleteSpecialProduct(db))

		apiv1.PUT("/products/:id", v1.EditProduct(db))

		apiv1.GET("/signIns", v1.GetSignIns(db))
		apiv1.GET("/signIns/:id", v1.GetSignIn(db))
		apiv1.PUT("/signIns/:id", v1.EditSignIn(db))

		apiv1.GET("/signIns/switch", v1.GetSignInSwitch(db))
		apiv1.PUT("/signIns/switch", v1.SetSignInSwitch(db))
	}

	return r
}
