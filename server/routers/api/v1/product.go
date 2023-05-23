package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strconv"
)

// @Summary Get a single product
// @Tags product
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} Response
// @Router /api/v1/products/{id} [get]
func GetProduct(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var item models.Product
		res := db.Model(&models.Product{}).Where("id = ?", id).Preload("Currency").Find(&item)
		if err = res.Error; err != nil {
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

// @Summary Get multiple products
// @Tags product
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Param   p_type    query    int     false        "type"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/products [get]
func GetProducts(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.Product{}).Order("id desc").Count(&total).Offset(int(offset)).Limit(int(pageSize))
		if pType := c.Query("p_type"); len(pType) > 0 {
			tx = tx.Where("p_type = ?", pType)
		}

		var items []models.Product
		if err := tx.Preload("Currency").Find(&items).Error; err != nil {
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

type EditProductReq struct {
	Price decimal.Decimal `json:"price"`
}

// @Summary Update product
// @Tags product
// @Produce  json
// @Param id path int true "id"
// @Param data body EditProductReq true "data"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/products/{id} [put]
func EditProduct(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var req EditProductReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}
		if req.Price.LessThanOrEqual(decimal.Zero) {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid price"))
			return
		}

		res := db.Model(&models.Product{}).Where("id = ?", id).Updates(&models.Product{Price: req.Price})
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
	}
}
