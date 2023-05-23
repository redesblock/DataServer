package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strconv"
)

// @Summary Get a single special product
// @Tags special product
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} Response
// @Router /api/v1/special_products/{id} [get]
func GetSpecialProduct(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var item models.SpecialProduct
		res := db.Model(&models.SpecialProduct{}).Where("id = ?", id).Preload("Product").Preload("Product.Currency").Find(&item)
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

// @Summary Get multiple special_products
// @Tags special product
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Param   p_type    query    int     false        "type"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/special_products [get]
func GetSpecialProducts(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.SpecialProduct{}).Order("id desc").Count(&total).Offset(int(offset)).Limit(int(pageSize))
		if pType := c.Query("p_type"); len(pType) > 0 {
			tx = tx.Where("p_type = ?", pType)
		}

		var items []models.SpecialProduct
		if err := tx.Preload("Product").Preload("Product.Currency").Find(&items).Error; err != nil {
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

type AddSpecialProductReq struct {
	Name     string             `json:"name"`
	PType    models.ProductType `json:"type"`
	Quantity uint64             `json:"quantity"`
	Discount decimal.Decimal    `json:"discount"`
	Reserve  uint64             `json:"reserve"`
}

// @Summary Add special product
// @Tags special product
// @Produce  json
// @Param data body AddSpecialProductReq true "data"
// @Success 200 {object} Response
// @Router /api/v1/special_products [post]
func AddSpecialProduct(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req AddSpecialProductReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		var t models.Product
		if res := db.Where("p_type = ?", req.PType).Find(&t); res.Error != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, res.Error))
			return
		} else if res.RowsAffected == 0 {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, "invalid type"))
			return
		}

		item := &models.SpecialProduct{
			Name:      req.Name,
			PType:     req.PType,
			Quantity:  req.Quantity,
			Discount:  req.Discount,
			Reserve:   req.Reserve,
			ProductID: t.ID,
		}
		res := db.Model(&models.SpecialProduct{}).Save(item)
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, item))
	}
}

type EditSpecialProductReq struct {
	Name    string `json:"name"`
	Reserve uint64 `json:"reserve"`
}

// @Summary Update special product
// @Tags special product
// @Produce  json
// @Param id path int true "id"
// @Param data body EditSpecialProductReq true "data"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/special_products/{id} [put]
func EditSpecialProduct(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var req EditSpecialProductReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		res := db.Model(&models.SpecialProduct{}).Where("id = ?", id).Updates(&models.SpecialProduct{
			Name:    req.Name,
			Reserve: req.Reserve,
		})
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
	}
}

// @Summary Delete article
// @Tags special product
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/special_products/{id} [delete]
func DeleteSpecialProduct(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		res := db.Unscoped().Where("id = ?", id).Delete(&models.SpecialProduct{})
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
	}

}
