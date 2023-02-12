package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"gorm.io/gorm"
	"net/http"
	"time"
)

// @Summary list traffics
// @Schemes
// @Description pagination list traffics
// @Tags report
// @Accept json
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} dataservice.ReportTraffic
// @Router /traffics [get]
func GetReportTrafficsHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize

		total, items, err := db.FindTraffics(offset, pageSize)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		pageTotal := total / pageSize
		if total%pageSize != 0 {
			pageTotal++
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, &List{
			Total:     total,
			PageTotal: pageTotal,
			Items:     items,
		}))
	}
}

// @Summary traffic info
// @Schemes
// @Description traffic info
// @Security ApiKeyAuth
// @Tags bucket
// @Accept json
// @Produce json
// @Param   date     path    int     true        "date"
// @Success 200 {object} dataservice.Bucket
// @Router /buckets/{date} [get]
func GetReportTrafficHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		start, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		var items []*dataservice.ReportTraffic
		err = db.Model(&dataservice.ReportTraffic{}).Order("nat_addr, timestamp ASC").Where("timestamp >= ? AND timestamp < ?", start.Unix(), start.Add(time.Hour*24).Unix()).Find(&items).Error
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err.Error()))
			return
		}
		//csv_upload := ""
		//csv_download := ""
		//nat_addr := ""
		//uploaded := int64(0)
		//uploadedCnt := int64(0)
		//downloaded := int64(0)
		//downloadedCnt := int64(0)
		//i := 0
		//for _, item := range items {
		//	t := time.Unix(item.Timestamp, 0)
		//	if nat_addr != item.NATAddr {
		//		nat_addr = item.NATAddr
		//		csv_upload += fmt.Sprintf("\n%s", nat_addr)
		//		csv_download += fmt.Sprintf("\n%s", nat_addr)
		//	}
		//	for i < t.Hour() {
		//
		//	}
		//	uploaded += item.Uploaded
		//	uploadedCnt += item.UploadedCnt
		//	downloaded += item.Downloaded
		//	downloadedCnt += item.DownloadedCnt
		//	csv_download += fmt.Sprintf(",%s")
		//	csv_upload += fmt.Sprintf(",%s")
		//}

		c.JSON(http.StatusOK, NewResponse(OKCode, items))
	}
}

type ReportTrafficReq struct {
	Timestamp     int64            `json:"timestamp"`
	Address       string           `json:"address"`
	Uploaded      map[string]int64 `json:"uploaded"`
	Downloaded    map[string]int64 `json:"downloaded"`
	UploadedCnt   map[string]int64 `json:"uploaded_cnt"`
	DownloadedCnt map[string]int64 `json:"downloaded_cnt"`
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
				traffic.UploadedCnt += req.UploadedCnt[key]
			}

			for key, size := range req.Downloaded {
				traffic, err := getItemFunc(key, req.NATAddr, req.Timestamp)
				if err != nil {
					continue
				}
				traffic.Downloaded += size
				traffic.DownloadedCnt += req.DownloadedCnt[key]
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
