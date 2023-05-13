package v1

import (
	"fmt"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type OverView struct {
	Buckets         uint64 `json:"buckets"`
	Objects         uint64 `json:"objects"`
	UsedStorage     uint64 `json:"used_storage"`
	TotalStorage    uint64 `json:"total_storage"`
	UsedStorageStr  string `json:"used_storage_str"`
	TotalStorageStr string `json:"total_storage_str"`
	UsedTraffic     uint64 `json:"used_traffic"`
	TotalTraffic    uint64 `json:"total_traffic"`
	UsedTrafficStr  string `json:"used_traffic_str"`
	TotalTrafficStr string `json:"total_traffic_str"`
}

// @Summary overview
// @Schemes
// @Description overview
// @Security ApiKeyAuth
// @Tags dashboard
// @Accept json
// @Produce json
// @Success 200 {object} OverView
// @Router /api/v1/overview [get]
func OverViewHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		item := &OverView{}

		userID, _ := c.Get("id")
		if err := db.Model(&models.Bucket{}).Where("user_id = ?", userID).Select("COUNT(id) AS count").Scan(&item.Buckets).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		if err := db.Model(&models.BucketObject{}).Where("status > ?", models.STATUS_WAIT).Where("bucket_id IN (?)", db.Model(&models.Bucket{}).Select("id").Where("user_id = ?", userID)).Select("COUNT(id) AS count").Scan(&item.Objects).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		if err := db.Model(&models.UsedStorage{}).Where("user_id = ?", userID).Select("COALESCE(SUM(num),0) AS total").Scan(&item.UsedStorage).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		if err := db.Model(&models.UsedTraffic{}).Where("user_id = ?", userID).Select("COALESCE(SUM(num),0) AS total").Scan(&item.UsedTraffic).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		var usr models.User
		if err := db.Model(&models.User{}).Where("id = ?", userID).Find(&usr).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		item.TotalTraffic = usr.TotalTraffic
		item.TotalTrafficStr = usr.TotalTrafficStr
		item.TotalStorage = usr.TotalStorage
		item.TotalStorageStr = usr.TotalStorageStr

		item.UsedTrafficStr = models.ByteSize(item.UsedTraffic)
		item.UsedStorageStr = models.ByteSize(item.UsedStorage)

		//if err := db.Model(&models.BillStorage{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size),0) AS total").Scan(&rtBucket.Total).Error; err != nil {
		//	c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
		//	return
		//}

		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

// @Summary used storage
// @Schemes
// @Description used storage
// @Security ApiKeyAuth
// @Tags dashboard
// @Param   start     query    string     false     "start time"
// @Param   end    query    string     false        "end time"
// @Param   uint    query    string     false        "unit"
// @Accept json
// @Produce json
// @Success 200 {object} models.UsedStorage
// @Router /api/v1/daily/storage [get]
func DailyStorageHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")
		userRole, _ := c.Get("role")

		div := uint64(1024)
		switch unit := strings.ToLower(c.Query("uint")); unit {
		case "kb":
			div = 1024
		case "mb":
			div = 1024 * 1024
		case "gb":
			div = 1024 * 1024 * 1024
		}
		endTime := time.Now()
		if t := c.Query("end"); len(t) > 0 {
			end, err := time.Parse(models.TIME_FORMAT, t)
			if err != nil {
				c.JSON(http.StatusOK, NewResponse(RequestCode, fmt.Errorf("invalid time %s", t)))
				return
			}
			endTime = end
		}

		startTime := endTime.Add(-7 * 24 * time.Hour)
		if t := c.Query("start"); len(t) > 0 {
			start, err := time.Parse(models.TIME_FORMAT, t)
			if err != nil {
				c.JSON(http.StatusOK, NewResponse(RequestCode, fmt.Errorf("invalid time %s", t)))
				return
			}
			startTime = start
		}

		var items []*models.UsedStorage
		if userRole.(uint) == uint(models.UserRole_Oper) || userRole.(uint) == uint(models.UserRole_Admin) {
			type result struct {
				Timestamp int64
				Total     uint64
			}
			var rets []*result
			if err := db.Model(&models.ReportTraffic{}).Order("timestamp desc").Where("timestamp >= ? AND timestamp <= ?", startTime.Unix(), endTime.Unix()).Select("timestamp, sum(uploaded) as total").Group("timestamp").Find(&rets).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}
			c.JSON(http.StatusOK, NewResponse(OKCode, func() []*models.UsedStorage {
				cnt := len(rets)
				for ; cnt > 0; cnt-- {
					ret := rets[cnt-1]
					items = append(items, &models.UsedStorage{
						Num:    ret.Total,
						NumStr: ret.Total / div,
						Time:   time.Unix(ret.Timestamp, 0).Format(models.TIME_FORMAT),
					})
				}
				return items
			}()))
		} else {
			var items []*models.UsedStorage
			if ret := db.Where("user_id = ?", userID).Order("time desc").Where("time >= ? and time <= ?", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")).Find(&items); ret.Error != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
				return
			}
			c.JSON(http.StatusOK, NewResponse(OKCode, func() []*models.UsedStorage {
				cnt := len(items)
				for ; cnt > 0; cnt-- {
					item := items[cnt-1]
					items = append(items, &models.UsedStorage{
						Num:    item.Num,
						NumStr: item.Num / div,
						Time:   item.Time,
					})
				}
				return items
			}()))
		}
	}
}

// @Summary used storage
// @Schemes
// @Description used storage
// @Security ApiKeyAuth
// @Tags dashboard
// @Param   start     query    string     false     "start time"
// @Param   end    query    string     false        "end time"
// @Param   uint    query    string     false        "unit"
// @Accept json
// @Produce json
// @Success 200 {object} models.UsedTraffic
// @Router /api/v1/daily/traffic [get]
func DailyTrafficHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")
		userRole, _ := c.Get("role")

		div := uint64(1024)
		switch unit := strings.ToLower(c.Query("uint")); unit {
		case "kb":
			div = 1024
		case "mb":
			div = 1024 * 1024
		case "gb":
			div = 1024 * 1024 * 1024
		}
		endTime := time.Now()
		if t := c.Query("end"); len(t) > 0 {
			end, err := time.Parse(models.TIME_FORMAT, t)
			if err != nil {
				c.JSON(http.StatusOK, NewResponse(RequestCode, fmt.Errorf("invalid time %s", t)))
				return
			}
			endTime = end
		}

		startTime := endTime.Add(-7 * 24 * time.Hour)
		if t := c.Query("start"); len(t) > 0 {
			start, err := time.Parse(models.TIME_FORMAT, t)
			if err != nil {
				c.JSON(http.StatusOK, NewResponse(RequestCode, fmt.Errorf("invalid time %s", t)))
				return
			}
			startTime = start
		}

		var items []*models.UsedTraffic
		if userRole.(uint) == uint(models.UserRole_Oper) || userRole.(uint) == uint(models.UserRole_Admin) {
			type result struct {
				Timestamp int64
				Total     uint64
			}
			var rets []*result
			if err := db.Model(&models.ReportTraffic{}).Order("timestamp desc").Where("timestamp >= ? AND timestamp <= ?", startTime.Unix(), endTime.Unix()).Select("timestamp, sum(downloaded) as total").Group("timestamp").Limit(24).Find(&rets).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}
			c.JSON(http.StatusOK, NewResponse(OKCode, func() []*models.UsedTraffic {
				cnt := len(rets)
				for ; cnt > 0; cnt-- {
					ret := rets[cnt-1]
					items = append(items, &models.UsedTraffic{
						Num:    ret.Total,
						NumStr: ret.Total / div,
						Time:   time.Unix(ret.Timestamp, 0).Format(models.TIME_FORMAT),
					})
				}
				return items
			}()))
		} else {
			var items []*models.UsedTraffic
			if ret := db.Where("user_id = ?", userID).Order("time desc").Where("time >= ? and time <= ?", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")).Find(&items); ret.Error != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
				return
			}
			c.JSON(http.StatusOK, NewResponse(OKCode, func() []*models.UsedTraffic {
				cnt := len(items)
				for ; cnt > 0; cnt-- {
					item := items[cnt-1]
					items = append(items, &models.UsedTraffic{
						Num:    item.Num,
						NumStr: item.Num / div,
						Time:   item.Time,
					})
				}
				return items
			}()))
		}
	}
}
