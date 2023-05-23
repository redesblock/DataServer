package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// @Summary list bucket objects
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Param   id     path    int     true        "bucket id"
// @Param   fid     query    int     false        "folder id"
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/buckets/{id}/objects [get]
func GetBucketObjectsHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}

		fid, err := strconv.ParseUint(c.DefaultQuery("fid", "0"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid fid"))
			return
		}

		total, items, err := models.FindBucketObjects(db, uint(id), uint(fid), offset, pageSize)
		if err != nil {
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

// @Summary bucket object info
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Produce json
// @Param   id     path    int     true        "bucket id"
// @Param   fid    path    int     true        "folder id"
// @Success 200 {object} Response
// @Router /api/v1/buckets/{id}/objects/{fid} [get]
func GetBucketObjectHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}

		fid, err := strconv.ParseUint(c.Param("fid"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid fid"))
			return
		}

		item, err := models.FindBucketObject(db, uint(id), uint(fid))
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}

		c.JSON(OKCode, NewResponse(c, OKCode, item))
	}
}

// @Summary remove bucket object
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Produce json
// @Param   id     path    int     true     "bucket id"
// @Param   fid     path    int     true        "folder id"
// @Success 200 {object} Response
// @Router /api/v1/buckets/{id}/objects/{fid} [delete]
func DeleteBucketObjectHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}

		fid, err := strconv.ParseUint(c.Param("fid"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid fid"))
			return
		}

		if err := db.Where("bucket_id = ?", id).Where("id = ?", fid).Delete(&models.BucketObject{}).Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}

		c.JSON(OKCode, NewResponse(c, OKCode, "ok"))
	}
}

type AddBucketReq struct {
	Parent uint   `json:"fid"`
	CID    string `json:"cid"`
}

// @Summary add bucket object
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Produce json
// @Param   id     path    int     true        "bucket id"
// @Param   fid     path    int    false     "folder id"
// @Param   name   path    string  true        "name"
// @Param   cid     path    string    false     "cid"
// @Param object body AddBucketReq false "object info"
// @Success 200 {object} Response
// @Router /api/v1/buckets/{id}/objects/{name} [post]
func AddBucketObjectHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var req AddBucketReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid fid"))
			return
		}
		fid := req.Parent

		name := c.Param("name")
		cid := req.CID

		var t *models.BucketObject
		if ret := db.Model(&models.BucketObject{}).Where("bucket_id = ?", id).Where("parent_id = ?", fid).Where("name = ?", name).Find(&t); ret.Error != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, ret.Error))
			return
		} else if ret.RowsAffected > 0 {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, "object already exist"))
			return
		}

		var item = &models.BucketObject{
			Name:     name,
			CID:      cid,
			ParentID: uint(fid),
			BucketID: uint(id),
		}
		if len(item.CID) > 0 {
			item.Status = models.STATUS_PINED
			response, err := http.Get(viper.GetString("gateway") + "/mop/" + cid + "/")
			if err == nil {
				size, _ := strconv.ParseUint(response.Header.Get("Decompressed-Content-Length"), 10, 64)
				item.Size = size
			} else {
				item.Status = models.STATUS_FAIL_PINED
			}
			item.UplinkProgress = 100
		}
		if err := db.Save(item).Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, item))
	}
}
