package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"
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
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var item models.Node
		res := db.Model(&models.Node{}).Where("id = ?", id).Find(&item)
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		if res.RowsAffected > 0 {
			c.JSON(http.StatusOK, NewResponse(OKCode, &item))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, nil))
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
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
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
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
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
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var req EditNodeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
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
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, res.RowsAffected > 0))
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
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		res := db.Unscoped().Where("id = ?", id).Delete(&models.Node{})
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, res.RowsAffected > 0))
	}

}
