package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// @Summary Get a single currency
// @Tags currency
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} Response
// @Router /api/v1/currencies/{id} [get]
func GetCurrency(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var item models.Currency
		res := db.Model(&models.Currency{}).Where("id = ?", id).Find(&item)
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

// @Summary Get multiple currencies
// @Tags currency
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/currencies [get]
func GetCurrencies(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.Currency{}).Order("id desc").Count(&total).Offset(int(offset)).Limit(int(pageSize))

		var items []models.Currency
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

type EditCurrencyReq struct {
	Rate decimal.Decimal `json:"rate"`
}

// @Summary Update currency
// @Tags currency
// @Produce  json
// @Param id path int true "id"
// @Param data body EditCurrencyReq true "data"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/currencies/{id} [put]
func EditCurrency(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var req EditCurrencyReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		if req.Rate.LessThanOrEqual(decimal.Zero) {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid rate"))
			return
		}

		res := db.Model(&models.Currency{}).Where("id = ?", id).Updates(&models.Currency{Rate: req.Rate})
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, res.RowsAffected > 0))
	}
}

// @Summary Delete article
// @Tags currency
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/currencies/{id} [delete]
func DeleteCurrency(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		res := db.Unscoped().Where("id = ? AND base = false", id).Delete(&models.Currency{})
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, res.RowsAffected > 0))
	}

}
