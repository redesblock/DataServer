package routers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
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
// @Router /overview [get]
func OverViewHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		item := &OverView{}

		userID, _ := c.Get("id")
		if err := db.Model(&dataservice.Bucket{}).Where("user_id = ?", userID).Select("COUNT(id) AS count").Scan(&item.Buckets).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		if err := db.Model(&dataservice.BucketObject{}).Where("status > ?", dataservice.STATUS_WAIT).Where("bucket_id IN (?)", db.Model(&dataservice.Bucket{}).Select("id").Where("user_id = ?", userID)).Select("COUNT(id) AS count").Scan(&item.Objects).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		if err := db.Model(&dataservice.UsedStorage{}).Where("user_id = ?", userID).Select("COALESCE(SUM(num),0) AS total").Scan(&item.UsedStorage).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		if err := db.Model(&dataservice.UsedTraffic{}).Where("user_id = ?", userID).Select("COALESCE(SUM(num),0) AS total").Scan(&item.UsedTraffic).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		var usr dataservice.User
		if err := db.Model(&dataservice.User{}).Where("id = ?", userID).Find(&usr).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		item.TotalTraffic = usr.TotalTraffic
		item.TotalTrafficStr = usr.TotalTrafficStr
		item.TotalStorage = usr.TotalStorage
		item.TotalStorageStr = usr.TotalStorageStr

		item.UsedTrafficStr = dataservice.ByteSize(item.UsedTraffic)
		item.UsedStorageStr = dataservice.ByteSize(item.UsedStorage)

		//if err := db.Model(&dataservice.BillStorage{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size),0) AS total").Scan(&rtBucket.Total).Error; err != nil {
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
// @Success 200 {object} dataservice.UsedStorage
// @Router /overview [get]
func DailyStorageHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")
		var items []*dataservice.UsedStorage
		date := time.Now()
		for i := 6; i >= 0; i-- {
			d := date.Add(-time.Hour * 24 * time.Duration(i))
			var item *dataservice.UsedStorage
			if ret := db.Where("user_id = ?", userID).Where("time = ?", d.Format("2006-01-02")).Find(&item); ret.Error != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
				return
			} else if ret.RowsAffected == 0 {
				item = &dataservice.UsedStorage{
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
		//var items []*dataservice.UsedStorage
		//if err := db.Model(&dataservice.UsedStorage{}).Order("id DESC").Where("user_id = ?", userID).Where("num > 0").Where("time > ?", before).Offset(int(offset)).Limit(int(limit)).Find(&items).Error; err != nil {
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
// @Success 200 {object} dataservice.UsedTraffic
// @Router /overview [get]
func DailyTrafficHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")

		var items []*dataservice.UsedTraffic
		date := time.Now()
		for i := 6; i >= 0; i-- {
			d := date.Add(-time.Hour * 24 * time.Duration(i))
			var item *dataservice.UsedTraffic
			if ret := db.Where("user_id = ?", userID).Where("time = ?", d.Format("2006-01-02")).Find(&item); ret.Error != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
				return
			} else if ret.RowsAffected == 0 {
				item = &dataservice.UsedTraffic{
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
		//var items []*dataservice.UsedTraffic
		//if err := db.Model(&dataservice.UsedTraffic{}).Order("id DESC").Where("user_id = ?", userID).Where("num > 0").Where("time > ?", before).Offset(int(offset)).Limit(int(limit)).Find(&items).Error; err != nil {
		//	c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
		//	return
		//}
		c.JSON(http.StatusOK, NewResponse(OKCode, items))
	}
}

// @Summary list user actions
// @Schemes
// @Description pagination query user actions
// @Security ApiKeyAuth
// @Tags dashboard
// @Accept json
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} dataservice.UserAction
// @Router /user/actions [get]
func UserActionsHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize

		userID, _ := c.Get("id")
		total, items, err := db.FindUserActions(userID.(uint), offset, pageSize)
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
