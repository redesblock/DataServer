package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"net/http"
)

type OverView struct {
	Buckets         uint64 `json:"buckets"`
	Objects         uint64 `json:"objects"`
	UsedStorage     uint64 `json:"used_storage"`
	TotalStorage    uint64 `json:"total_storage"`
	UsedStorageStr  string `json:"used_storage_str"`
	TotalStorageStr string `json:"total_storage_str"`
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
		type Result struct {
			Count uint64
			Total uint64
		}

		userID, _ := c.Get("id")
		var rtBucket Result
		if err := db.Model(&dataservice.Bucket{}).Where("user_id = ?", userID).Select("COUNT(id) AS count").Scan(&rtBucket.Count).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		var rtBucketObject Result
		if err := db.Model(&dataservice.BucketObject{}).Where("c_id != ''").Where("bucket_id IN (?)", db.Model(&dataservice.Bucket{}).Select("id").Where("user_id = ?", userID)).Select("COALESCE(SUM(size),0) AS total,COUNT(id) AS count").Scan(&rtBucketObject).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		if err := db.Model(&dataservice.BillStorage{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size),0) AS total").Scan(&rtBucket.Total).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, &OverView{
			Buckets:         rtBucket.Count,
			Objects:         rtBucketObject.Count,
			TotalStorage:    rtBucket.Total,
			UsedStorage:     rtBucketObject.Total,
			TotalStorageStr: dataservice.HumanateBytes(rtBucket.Total),
			UsedStorageStr:  dataservice.HumanateBytes(rtBucketObject.Total),
		}))
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

		offset := 0
		limit := 10
		var items []*dataservice.UsedStorage
		if err := db.Model(&dataservice.UsedStorage{}).Order("id DESC").Where("user_id = ?", userID).Where("num > 0").Offset(int(offset)).Limit(int(limit)).Find(&items).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
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

		offset := 0
		limit := 10
		var items []*dataservice.UsedTraffic
		if err := db.Model(&dataservice.UsedTraffic{}).Order("id DESC").Where("user_id = ?", userID).Where("num > 0").Offset(int(offset)).Limit(int(limit)).Find(&items).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
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
