package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"gorm.io/gorm"
	"net/http"
)

type ReportTrafficReq struct {
	Timestamp     int64            `json:"timestamp"`
	Address       string           `json:"address"`
	Uploaded      map[string]int64 `json:"uploaded"`
	Downloaded    map[string]int64 `json:"downloaded"`
	UploadedCnt   int64            `json:"uploaded_cnt"`
	DownloadedCnt int64            `json:"downloaded_cnt"`
	Signed        string           `json:"signed"`
	NATAddr       string           `json:"nat_addr"`
}

// @Summary report traffic
// @Schemes
// @Description add report traffic
// @Tags report
// @Accept json
// @Produce json
// @Success 200 {string}
// @Router /traffic [post]
func AddReportTrafficHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req ReportTrafficReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			items := make(map[string]*dataservice.ReportTraffic)
			getItemFunc := func(key, nat_addr string, timestamp int64) (*dataservice.ReportTraffic, error) {
				k := fmt.Sprintf("%s_%d", key, timestamp)
				if item, ok := items[k]; ok {
					return item, nil
				}
				var item dataservice.ReportTraffic
				if result := db.Find(&item, "token = ? AND timestamp =? AND nat_addr = ?", key, timestamp, nat_addr); result.Error != nil {
					return nil, result.Error
				} else if result.RowsAffected == 0 {
					item.Token = key
					item.Timestamp = timestamp
					item.NATAddr = nat_addr
					items[k] = &item
				}
				return &item, nil
			}

			for key, size := range req.Uploaded {
				traffic, err := getItemFunc(key, req.NATAddr, req.Timestamp)
				if err != nil {
					continue
				}
				traffic.Uploaded += size
			}

			for key, size := range req.Downloaded {
				traffic, err := getItemFunc(key, req.NATAddr, req.Timestamp)
				if err != nil {
					continue
				}
				traffic.Downloaded += size
			}

			if cnt := len(items); cnt > 0 {
				traffics := make([]*dataservice.ReportTraffic, cnt)
				i := 0
				for _, item := range items {
					traffics[i] = item
					i++
				}
				return tx.Save(traffics).Error
			}
			return nil
		}); err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err.Error()))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, "ok"))
	}
}

// @Summary report receipt
// @Schemes
// @Description add report receipt
// @Tags report
// @Accept json
// @Produce json
// @Success 200 {string}
// @Router /receipt [post]
func AddReportReceiptHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, NewResponse(OKCode, "ok"))
	}
}
