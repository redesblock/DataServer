package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

// @Summary Get a single node
// @Tags node
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} Response
// @Router /api/v1/nodes/{id} [get]
func GetNode(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var item models.Node
		res := db.Model(&models.Node{}).Where("id = ?", id).Find(&item)
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		if res.RowsAffected > 0 {
			c.JSON(OKCode, NewResponse(c, OKCode, &item))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, nil))
	}
}

// @Summary Get multiple nodes
// @Tags node
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/nodes [get]
func GetNodes(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.Node{}).Order("id desc").Count(&total).Offset(int(offset)).Limit(int(pageSize))

		var items []models.Node
		if err := tx.Find(&items).Error; err != nil {
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

type AddNodeReq struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Zone      string `json:"zone"`
	VoucherID string `json:"voucher_id"`
}

// @Summary Add node
// @Tags node
// @Produce  json
// @Param data body AddNodeReq true "data"
// @Success 200 {object} Response
// @Router /api/v1/nodes [post]
func AddNode(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req AddNodeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		item := &models.Node{
			Name:      req.Name,
			IP:        req.IP,
			Port:      1683,
			Zone:      req.Zone,
			VoucherID: req.VoucherID,
			Usable:    true,
		}
		res := db.Model(&models.Node{}).Save(item)
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, item))
	}
}

type EditNodeReq struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	//Port      int    `json:"port"`
	Zone      string `json:"zone"`
	VoucherID string `json:"voucher_id"`
}

// @Summary Update node
// @Tags node
// @Produce  json
// @Param id path int true "id"
// @Param data body EditNodeReq true "data"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/nodes/{id} [put]
func EditNode(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var req EditNodeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		res := db.Model(&models.Node{}).Where("id = ?", id).Updates(&models.Node{
			Name:      req.Name,
			IP:        req.IP,
			Zone:      req.Zone,
			VoucherID: req.VoucherID,
			Usable:    true,
		})
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
	}
}

// @Summary Delete article
// @Tags node
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/nodes/{id} [delete]
func DeleteNode(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		res := db.Unscoped().Where("id = ?", id).Delete(&models.Node{})
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
	}

}

func GetAreasHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var items []string
		err := db.Model(&models.Node{}).Select("zone").Where("usable = true").Order("zone DESC").Find(&items).Error
		if err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, items))
	}
}

func GetNetWorksHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(OKCode, NewResponse(c, OKCode, strings.Split("MOP Storage", ",")))
	}
}

// @Summary used storage
// @Schemes
// @Description used storage
// @Security ApiKeyAuth
// @Tags dashboard
// @Param   start     query    string     false     "start time"
// @Param   end    query    string     false        "end time"
// @Param   unit    query    string     false        "unit"
// @Accept json
// @Produce json
// @Success 200 {object} models.UsedStorage
// @Router /api/v1/nodes/storage [get]
func NodeStorageHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		div := uint64(1024)
		switch unit := strings.ToLower(c.Query("unit")); unit {
		case "kb":
			div = 1024
		case "mb":
			div = 1024 * 1024
		case "gb":
			div = 1024 * 1024 * 1024
		}
		endTime := time.Now()
		if t := c.Query("end"); len(t) > 0 {
			end, err := time.Parse("2006-01-02", t)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid time %s", t)))
				return
			}
			endTime = end
		}

		startTime := endTime.Add(-7 * 24 * time.Hour)
		if t := c.Query("start"); len(t) > 0 {
			start, err := time.Parse("2006-01-02", t)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid time %s", t)))
				return
			}
			startTime = start
		}

		var items []*models.UsedStorage
		type result struct {
			Timestamp int64
			Total     uint64
		}
		var rets []*result
		if err := db.Model(&models.ReportTraffic{}).Order("timestamp desc").Where("timestamp >= ? AND timestamp <= ?", startTime.Unix(), endTime.Unix()).Select("timestamp, sum(uploaded) as total").Group("timestamp").Find(&rets).Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, func() []*models.UsedStorage {
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
	}
}

// @Summary used storage
// @Schemes
// @Description used storage
// @Security ApiKeyAuth
// @Tags dashboard
// @Param   start     query    string     false     "start time"
// @Param   end    query    string     false        "end time"
// @Param   unit    query    string     false        "unit"
// @Accept json
// @Produce json
// @Success 200 {object} models.UsedTraffic
// @Router /api/v1/nodes/traffic [get]
func NodeTrafficHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		div := uint64(1024)
		switch unit := strings.ToLower(c.Query("unit")); unit {
		case "kb":
			div = 1024
		case "mb":
			div = 1024 * 1024
		case "gb":
			div = 1024 * 1024 * 1024
		}
		endTime := time.Now()
		if t := c.Query("end"); len(t) > 0 {
			end, err := time.Parse("2006-01-02", t)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid time %s", t)))
				return
			}
			endTime = end
		}

		startTime := endTime.Add(-7 * 24 * time.Hour)
		if t := c.Query("start"); len(t) > 0 {
			start, err := time.Parse("2006-01-02", t)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid time %s", t)))
				return
			}
			startTime = start
		}

		var items []*models.UsedTraffic
		type result struct {
			Timestamp int64
			Total     uint64
		}
		var rets []*result
		if err := db.Model(&models.ReportTraffic{}).Order("timestamp desc").Where("timestamp >= ? AND timestamp <= ?", startTime.Unix(), endTime.Unix()).Select("timestamp, sum(downloaded) as total").Group("timestamp").Limit(24).Find(&rets).Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, func() []*models.UsedTraffic {
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
	}
}
