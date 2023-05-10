package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// @Summary Get multiple buckets
// @Tags bucket
// @Security ApiKeyAuth
// @Accept json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/buckets [get]
func GetBucketsHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")

		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		total, items, err := models.FindBuckets(db, userID.(uint), offset, pageSize)
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

// @Summary Get a single bucket
// @Tags bucket
// @Security ApiKeyAuth
// @Accept json
// @Param   id     path    int     true        "bucket id"
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/buckets/{id} [get]
func GetBucketHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}

		userID, _ := c.Get("id")
		item, err := models.FindBucket(db, userID.(uint), uint(id))
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

// @Summary remove bucket
// @Security ApiKeyAuth
// @Tags bucket
// @Accept json
// @Param   id     path    int     true        "bucket id"
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/buckets/{id} [delete]
func DeleteBucketHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}

		userID, _ := c.Get("id")
		if err := db.Where("user_id = ?", userID).Where("id = ?", id).Delete(&models.Bucket{}).Error; err != nil {
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
// @Security ApiKeyAuth
// @Tags bucket
// @Accept json
// @Param bucket body Bucket true "bucket info"
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/buckets [post]
func AddBucketHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req Bucket
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		userID, _ := c.Get("id")
		var item *models.Bucket
		if ret := db.Model(&models.Bucket{}).Where("user_id = ?", userID).Where("name = ?", req.Name).Find(&item); ret.Error != nil {
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
// @Security ApiKeyAuth
// @Tags bucket
// @Accept json
// @Produce json
// @Param   id     path    int     true        "bucket id"
// @Param bucket body Bucket true "update bucket info"
// @Success 200 {object} Response
// @Router /api/v1/buckets/{id} [post]
func UpdateBucketHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req Bucket
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		userID, _ := c.Get("id")
		var item *models.Bucket
		if ret := db.Model(&models.Bucket{}).Where("user_id = ?", userID).Where("id = ?", c.Param("id")).Find(&item); ret.Error != nil {
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
