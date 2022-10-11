package main

import (
	"github.com/redesblock/dataserver/dataservice"
	"github.com/redesblock/dataserver/server"
	"os"
)

// @title DataServer Swagger API
// @version 1.0
// @description This is a sample server DataServer server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 127.0.0.1:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	mode, ok := os.LookupEnv("DATA_SERVER_DB_MODE")
	if !ok {
		mode = "sqlite"
	}
	dsn, ok := os.LookupEnv("DATA_SERVER_DB_DSN")
	if !ok {
		dsn = "gateway.db"
	}
	port, ok := os.LookupEnv("DATA_SERVER_PORT")
	if !ok {
		port = "8080"
	}
	db := dataservice.New(mode, dsn)
	server.Start(":"+port, db)
}

// DATA_SERVER_DB_MODE
// DATA_SERVER_DB_DSN
// DATA_SERVER_PORT
// DATA_SERVER_JWT_SECRET
// DATA_SERVER_AREA
// DATA_SERVER_NETWORK
// DATA_SERVER_TRAFFIC_PRICE
// DATA_SERVER_STORAGE_PRICE
// DATA_SERVER_GATEWAY
// DATA_SERVER_MOP
// DATA_SERVER_VOUCHER
