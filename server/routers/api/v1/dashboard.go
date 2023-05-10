package v1

import (
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"net/http"
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
// @Accept json
// @Produce json
// @Success 200 {object} models.UsedStorage
// @Router /api/v1/daily/storage [get]
func DailyStorageHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")
		var items []*models.UsedStorage
		if userID.(uint) == 100 {
			type result struct {
				Timestamp int64
				Total     uint64
			}
			var rets []*result
			if err := db.Model(&models.ReportTraffic{}).Order("timestamp desc").Select("timestamp, sum(uploaded) as total").Group("timestamp").Limit(24).Find(&rets).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}
			c.JSON(http.StatusOK, NewResponse(OKCode, func() []*models.UsedStorage {
				cnt := len(rets)
				for ; cnt > 0; cnt-- {
					ret := rets[cnt-1]
					items = append(items, &models.UsedStorage{
						Num:    ret.Total,
						NumStr: ret.Total / 1024,
						Time:   time.Unix(ret.Timestamp, 0).Format(models.TIME_FORMAT),
					})
				}
				return items
			}()))
			return
		}
		date := time.Now()
		for i := 6; i >= 0; i-- {
			d := date.Add(-time.Hour * 24 * time.Duration(i))
			var item *models.UsedStorage
			if ret := db.Where("user_id = ?", userID).Where("time = ?", d.Format("2006-01-02")).Find(&item); ret.Error != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
				return
			} else if ret.RowsAffected == 0 {
				item = &models.UsedStorage{
					UserID: userID.(uint),
					Time:   d.Format("2006-01-02"),
				}
			}
			items = append(items, item)
		}

		//offset := 0
		//limit := 10
		//
		//before := time.Now().Add(-time.Hour * 24 * 7).Format("2006-01-02")
		//var items []*models.UsedStorage
		//if err := db.Model(&models.UsedStorage{}).Order("id DESC").Where("user_id = ?", userID).Where("num > 0").Where("time > ?", before).Offset(int(offset)).Limit(int(limit)).Find(&items).Error; err != nil {
		//	c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
		//	return
		//}
		c.JSON(http.StatusOK, NewResponse(OKCode, items))
	}
}

// @Summary used storage
// @Schemes
// @Description used storage
// @Security ApiKeyAuth
// @Tags dashboard
// @Accept json
// @Produce json
// @Success 200 {object} models.UsedTraffic
// @Router /api/v1/daily/traffic [get]
func DailyTrafficHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")

		var items []*models.UsedTraffic
		if userID.(uint) == 100 {
			type result struct {
				Timestamp int64
				Total     uint64
			}
			var rets []*result
			if err := db.Model(&models.ReportTraffic{}).Order("timestamp desc").Select("timestamp, sum(downloaded) as total").Group("timestamp").Limit(24).Find(&rets).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}
			c.JSON(http.StatusOK, NewResponse(OKCode, func() []*models.UsedTraffic {
				cnt := len(rets)
				for ; cnt > 0; cnt-- {
					ret := rets[cnt-1]
					items = append(items, &models.UsedTraffic{
						Num:    ret.Total,
						NumStr: ret.Total / 1024,
						Time:   time.Unix(ret.Timestamp, 0).Format(models.TIME_FORMAT),
					})
				}
				return items
			}()))
			return
		}

		date := time.Now()
		for i := 6; i >= 0; i-- {
			d := date.Add(-time.Hour * 24 * time.Duration(i))
			var item *models.UsedTraffic
			if ret := db.Where("user_id = ?", userID).Where("time = ?", d.Format("2006-01-02")).Find(&item); ret.Error != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
				return
			} else if ret.RowsAffected == 0 {
				item = &models.UsedTraffic{
					UserID: userID.(uint),
					Time:   d.Format("2006-01-02"),
				}
			}
			items = append(items, item)
		}

		//offset := 0
		//limit := 10
		//
		//before := time.Now().Add(-time.Hour * 24 * 7).Format("2006-01-02")
		//var items []*models.UsedTraffic
		//if err := db.Model(&models.UsedTraffic{}).Order("id DESC").Where("user_id = ?", userID).Where("num > 0").Where("time > ?", before).Offset(int(offset)).Limit(int(limit)).Find(&items).Error; err != nil {
		//	c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
		//	return
		//}
		c.JSON(http.StatusOK, NewResponse(OKCode, items))
	}
}
