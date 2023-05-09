package main

import (
	"github.com/redesblock/dataserver/cmd"
	_ "github.com/redesblock/dataserver/docs"
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
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	cmd.Execute()
}
