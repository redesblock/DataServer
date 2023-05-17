package v1

import (
	"fmt"
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
		userRole, _ := c.Get("role")
		fmt.Println(userRole, userID)
		if userRole.(models.UserRole) == models.UserRole_Oper || userRole.(models.UserRole) == models.UserRole_Admin {
			if err := db.Model(&models.Bucket{}).Select("COUNT(id) AS count").Scan(&item.Buckets).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}

			if err := db.Model(&models.BucketObject{}).Where("status > ?", models.STATUS_WAIT).Select("COUNT(id) AS count").Scan(&item.Objects).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}

			if err := db.Model(&models.UsedStorage{}).Select("COALESCE(SUM(num),0) AS total").Scan(&item.UsedStorage).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}

			if err := db.Model(&models.UsedTraffic{}).Select("COALESCE(SUM(num),0) AS total").Scan(&item.UsedTraffic).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}

			if err := db.Model(&models.User{}).Select("COALESCE(SUM(total_storage),0) AS total_storage, COALESCE(SUM(total_traffic),0) AS total_traffic").Scan(&item).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}

			var usr models.User
			if err := db.Model(&models.User{}).Where("id = ?", userID).Find(&usr).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}
			item.TotalTrafficStr = models.ByteSize(item.TotalTraffic)
			item.TotalStorageStr = models.ByteSize(item.TotalStorage)

			item.UsedTrafficStr = models.ByteSize(item.UsedTraffic)
			item.UsedStorageStr = models.ByteSize(item.UsedStorage)

			c.JSON(http.StatusOK, NewResponse(OKCode, item))
			return
		}
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
		userRole, _ := c.Get("role")

		var items []*models.UsedStorage
		date := time.Now()
		if userRole.(models.UserRole) == models.UserRole_Oper || userRole.(models.UserRole) == models.UserRole_Admin {
			for i := 6; i >= 0; i-- {
				d := date.Add(-time.Hour * 24 * time.Duration(i))
				var item models.UsedStorage
				item.Time = d.Format("2006-01-02")
				if ret := db.Where("time = ?", item.Time).Select("COALESCE(SUM(num),0) AS num").Find(&item); ret.Error != nil {
					c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
					return
				}
				items = append(items, &item)
			}
			c.JSON(http.StatusOK, NewResponse(OKCode, items))
			return
		}
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
			c.JSON(http.StatusOK, NewResponse(OKCode, items))
			return
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

		var items []*models.UsedTraffic
		date := time.Now()
		if userRole.(models.UserRole) == models.UserRole_Oper || userRole.(models.UserRole) == models.UserRole_Admin {
			for i := 6; i >= 0; i-- {
				d := date.Add(-time.Hour * 24 * time.Duration(i))
				var item models.UsedTraffic
				item.Time = d.Format("2006-01-02")
				if ret := db.Where("time = ?", item.Time).Select("COALESCE(SUM(num),0) AS num").Find(&item); ret.Error != nil {
					c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
					return
				}
				items = append(items, &item)
			}
			c.JSON(http.StatusOK, NewResponse(OKCode, items))
			return
		}
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
			c.JSON(http.StatusOK, NewResponse(OKCode, items))
			return
		}
	}
}
