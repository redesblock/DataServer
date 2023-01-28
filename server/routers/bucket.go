package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"net/http"
	"strconv"
)

// @Summary list buckets
// @Schemes
// @Description pagination list buckets
// @Security ApiKeyAuth
// @Tags bucket
// @Accept json
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} dataservice.Bucket
// @Router /buckets [get]
func GetBucketsHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize

		userID, _ := c.Get("id")
		total, items, err := db.FindBuckets(userID.(uint), offset, pageSize)
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

// @Summary bucket info
// @Schemes
// @Description bucket info
// @Security ApiKeyAuth
// @Tags bucket
// @Accept json
// @Produce json
// @Param   id     path    int     true        "bucket id"
// @Success 200 {object} dataservice.Bucket
// @Router /buckets/{id} [get]
func GetBucketHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}

		userID, _ := c.Get("id")
		item, err := db.FindBucket(userID.(uint), uint(id))
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

// @Summary remove bucket
// @Schemes
// @Description remove bucket
// @Security ApiKeyAuth
// @Tags bucket
// @Accept json
// @Produce json
// @Param   id     path    int     true        "bucket id"
// @Success 200 {object} dataservice.Bucket
// @Router /buckets/{id} [delete]
func DeleteBucketHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}

		userID, _ := c.Get("id")
		if err := db.Where("user_id = ?", userID).Where("id = ?", id).Delete(&dataservice.Bucket{}).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, "ok"))
	}
}

type Bucket struct {
	Name    string `json:"name"`
	Access  bool   `json:"access"`
	Network string `json:"network"`
	Area    string `json:"area"`
}

// @Summary add bucket
// @Schemes
// @Description add bucket
// @Security ApiKeyAuth
// @Tags bucket
// @Accept json
// @Produce json
// @Param bucket body Bucket true "bucket info"
// @Success 200 {object} dataservice.Bucket
// @Router /buckets [post]
func AddBucketHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req Bucket
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		userID, _ := c.Get("id")
		var item *dataservice.Bucket
		if ret := db.Model(&dataservice.Bucket{}).Where("user_id = ?", userID).Where("name = ?", req.Name).Find(&item); ret.Error != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
			return
		} else if ret.RowsAffected > 0 {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, "bucket already exist"))
			return
		}

		item.Name = req.Name
		item.Area = req.Area
		item.Access = req.Access
		item.Network = req.Network
		item.UserID = userID.(uint)

		if err := db.Save(item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

// @Summary update bucket
// @Schemes
// @Description update bucket
// @Security ApiKeyAuth
// @Tags bucket
// @Accept json
// @Produce json
// @Param   id     path    int     true        "bucket id"
// @Param bucket body Bucket true "update bucket info"
// @Success 200 {object} dataservice.Bucket
// @Router /buckets/{id} [post]
func UpdateBucketHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req Bucket
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		userID, _ := c.Get("id")
		var item *dataservice.Bucket
		if ret := db.Model(&dataservice.Bucket{}).Where("user_id = ?", userID).Where("id = ?", c.Param("id")).Find(&item); ret.Error != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
			return
		} else if ret.RowsAffected == 0 {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, "not found"))
			return
		}

		item.Name = req.Name
		item.Area = req.Area
		item.Access = req.Access
		item.Network = req.Network
		item.UserID = userID.(uint)

		if err := db.Save(item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}
