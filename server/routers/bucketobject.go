package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"net/http"
	"strconv"
)

// @Summary list bucket objects
// @Schemes
// @Description pagination query bucket objects
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Produce json
// @Param   id     path    int     true        "bucket id"
// @Param   fid     query    int     false        "folder id"
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} dataservice.BucketObject
// @Router /buckets/{id}/objects [get]
func GetBucketObjectsHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}

		fid, err := strconv.ParseUint(c.DefaultQuery("fid", "0"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid fid"))
			return
		}

		total, items, err := db.FindBucketObjects(uint(id), uint(fid), offset, pageSize)
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

// @Summary bucket object info
// @Schemes
// @Description bucket object info
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Produce json
// @Param   id     path    int     true        "bucket id"
// @Param   fid    path    int     true        "folder id"
// @Success 200 {object} dataservice.BucketObject
// @Router /buckets/{id}/objects/{fid} [get]
func GetBucketObjectHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}

		fid, err := strconv.ParseUint(c.Param("fid"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid fid"))
			return
		}

		item, err := db.FindBucketObject(uint(id), uint(fid))
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

// @Summary remove bucket object
// @Schemes
// @Description remove bucket object
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Produce json
// @Param   id     path    int     true     "bucket id"
// @Param   fid     path    int     true        "folder id"
// @Success 200 {object} dataservice.BucketObject
// @Router /buckets/{id}/objects/{fid} [delete]
func DeleteBucketObjectHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}

		fid, err := strconv.ParseUint(c.Param("fid"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid fid"))
			return
		}

		if err := db.Where("bucket_id = ?", id).Where("id = ?", fid).Delete(&dataservice.BucketObject{}).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, "ok"))
	}
}

type AddBucketReq struct {
	Parent uint `json:"fid"`
}

// @Summary add bucket object
// @Schemes
// @Description add bucket folder
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Produce json
// @Param   id     path    int     true        "bucket id"
// @Param   fid     path    int    false     "folder id"
// @Param   name   path    string  true        "name"
// @Param   cid     path    string    false     "cid"
// @Param object body AddBucketReq false "object info"
// @Success 200 {object} dataservice.BucketObject
// @Router /buckets/{id}/objects/{name} [post]
func AddBucketObjectHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var req AddBucketReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid fid"))
			return
		}
		fid := req.Parent

		name := c.Param("name")
		cid := c.Query("cid")

		var t *dataservice.BucketObject
		if ret := db.Model(&dataservice.BucketObject{}).Where("bucket_id = ?", id).Where("parent_id = ?", fid).Where("name = ?", name).Find(&t); ret.Error != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
			return
		} else if ret.RowsAffected > 0 {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, "object already exist"))
			return
		}

		var item = &dataservice.BucketObject{
			Name:     name,
			CID:      cid,
			Status:   dataservice.STATUS_UPLOADED,
			ParentID: uint(fid),
			BucketID: uint(id),
		}
		if err := db.Save(item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}
