package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"net/http"
)

// @Summary report traffic
// @Schemes
// @Description add report traffic
// @Tags report
// @Accept json
// @Produce json
// @Success 200 {object} dataservice.ReportTraffic
// @Router /report/traffic [post]
func AddReportTrafficHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, NewResponse(OKCode, nil))
	}
}
