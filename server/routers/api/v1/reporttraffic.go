package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// @Summary Get multiple traffics
// @Schemes
// @Tags report traffic
// @Accept json
// @Produce json
// @Param   ip     query    string     false        "ip"
// @Param   start   query    string     true        "start"
// @Param   end   query    string     true        "end"
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Router /api/v1/traffics [get]
func GetReportTraffics(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.ReportTraffic{}).Order("timestamp DESC, nat_addr")

		start := c.Query("start")
		end := c.Query("end")
		if len(start) > 0 && len(end) > 0 {
			startTime, err := time.Parse("2006-01-02", start)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
				return
			}
			endTime, err := time.Parse("2006-01-02", end)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
				return
			}
			if startTime.After(endTime) {
				tx = tx.Where("timestamp >= ? AND timestamp < ?", endTime.Unix(), startTime.Unix())
			} else {
				tx = tx.Where("timestamp >= ? AND timestamp < ?", startTime.Unix(), endTime.Unix())
			}
		}
		if ip := c.Query("ip"); len(ip) > 0 {
			tx = tx.Where("nat_addr LIKE ?", fmt.Sprintf("%s%%", ip))
		}

		var items []*models.ReportTraffic
		if err := tx.Count(&total).Offset(int(offset)).Limit(int(pageSize)).Find(&items).Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}

		pageTotal := total / pageSize
		if total%pageSize != 0 {
			pageTotal++
		}
		c.JSON(OKCode, NewResponse(c, OKCode, &List{
			Total:     total,
			PageTotal: pageTotal,
			Items:     items,
		}))
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

// @Summary add report traffic
// @Schemes
// @Tags report traffic
// @Accept json
// @Produce json
// @Param data body ReportTrafficReq true "data"
// @Success 200 {object} Response
// @Router /api/v1/traffic [post]
func AddReportTraffic(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req ReportTrafficReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			items := make(map[string]*models.ReportTraffic)
			getItemFunc := func(key, nat_addr string, timestamp int64) (*models.ReportTraffic, error) {
				k := fmt.Sprintf("%s_%d", key, timestamp)
				if item, ok := items[k]; ok {
					return item, nil
				}
				var item models.ReportTraffic
				if result := tx.Find(&item, "token = ? AND timestamp =? AND nat_addr = ?", key, timestamp, nat_addr); result.Error != nil {
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
				traffics := make([]*models.ReportTraffic, cnt)
				i := 0
				for _, item := range items {
					traffics[i] = item
					i++
				}
				return tx.Save(traffics).Error
			}
			return nil
		}); err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err.Error()))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, "ok"))
	}
}

func GetBillReport(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		startTime := time.Now().Truncate(24 * time.Hour)
		if start := c.Query("start_time"); len(start) > 0 {
			t, err := time.Parse("2006-01-02", start)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
				return
			}
			startTime = t
		}
		endTime := startTime.Add(time.Hour * 24)
		if end := c.Query("end_time"); len(end) > 0 {
			t, err := time.Parse("2006-01-02", end)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
				return
			}
			endTime = t
		}

		tx := db.Model(&models.Order{}).Where("status=?", models.OrderSuccess).Where("note=?", "xb")

		if startTime.After(endTime) {
			tx = tx.Where("created_at >= ? AND created_at < ?", endTime.Format("2006-01-02"), startTime.Format("2006-01-02"))
		} else {
			tx = tx.Where("created_at >= ? AND created_at < ?", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
		}

		type Result struct {
			StartTime string  `json:"start_time"`
			EndTime   string  `json:"end_time"`
			RowCount  int64   `json:"row_count"`
			Amount    float64 `json:"amount"`
			RMBAmount float64 `json:"rmb_amount"`
		}

		result := &Result{
			StartTime: startTime.Format("2006-01-02"),
			EndTime:   endTime.Format("2006-01-02"),
		}
		err := tx.Select("SUM(CONVERT(price, DECIMAL(10,2))) AS rmb_amount, SUM(payment_amount) as amount, COUNT(*) as row_count").
			Scan(result).Error
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err.Error()))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, result))
	}
}

func GetBillEmail(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		n, err := strconv.ParseInt(c.DefaultQuery("num", "1"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}
		var data []string
		err = db.Model(&models.User{}).Select("email").Where("reserve=?", 1).Order("RAND()").Limit(int(n)).Find(&data).Error
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err.Error()))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, data))
	}
}
