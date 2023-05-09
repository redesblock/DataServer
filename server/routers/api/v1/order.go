package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// @Summary Get a single order
// @Tags order
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} Response
// @Router /api/v1/orders/{id} [get]
func GetOrder(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var item models.Order
		res := db.Model(&models.Order{}).Where("id = ?", id).Find(&item)
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

// @Summary Get multiple orders
// @Tags order
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/orders [get]
func GetOrders(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.Node{}).Order("id desc").Count(&total).Offset(int(offset)).Limit(int(pageSize))

		var items []models.Order
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

type AddOrderReq struct {
}

// @Summary Add order
// @Produce  json
// @Param data body AddOrderReq true "data"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/articles [post]
func AddOrder(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req AddOrderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		item := &models.Order{}
		res := db.Model(&models.Node{}).Save(item)
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

type EditOrderReq struct {
}

// @Summary Update order
// @Produce  json
// @Param id path int true "id"
// @Param data body EditOrderReq true "data"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/orders/{id} [put]
func EditOrder(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var req EditOrderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		res := db.Model(&models.Node{}).Where("id = ?", id).Updates(&models.Order{})
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, res.RowsAffected > 0))
	}
}
