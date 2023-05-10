package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
)

// @Summary list storage bills
// @Schemes
// @Description pagination query storage bills
// @Security ApiKeyAuth
// @Tags bills
// @Accept json
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Router /api/v1/bills/storage [get]
func GetBillsStorageHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		userID, _ := c.Get("id")

		var items []models.Order
		tx := db.Model(&models.Order{}).Order("id desc").Where("user_id = ?", userID).Where("p_type = ?", models.ProductType_Storage)
		if err := tx.Count(&total).Offset(int(offset)).Limit(int(pageSize)).Find(&items).Error; err != nil {
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

// @Summary list traffic bills
// @Schemes
// @Description pagination query traffic bills
// @Tags bills
// @Accept json
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Router /api/v1/bills/traffic [get]
func GetBillsTrafficHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		userID, _ := c.Get("id")

		var items []models.Order
		tx := db.Model(&models.Order{}).Order("id desc").Where("user_id = ?", userID).Where("p_type = ?", models.ProductType_Traffic)
		if err := tx.Count(&total).Offset(int(offset)).Limit(int(pageSize)).Find(&items).Error; err != nil {
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

type BillReq struct {
	Hash        string          `json:"hash"`
	Amount      string          `json:"amount"`
	Size        decimal.Decimal `json:"size"`
	Description string          `json:"description"`
}

// @Summary add storage bill
// @Schemes
// @Description add storage bill
// @Security ApiKeyAuth
// @Tags bills
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/bills/storage [post]
func AddBillsStorageHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req BillReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		userID, _ := c.Get("id")
		item := &models.Order{
			PType:    models.ProductType_Storage,
			Hash:     req.Hash,
			Quantity: req.Size.Mul(decimal.NewFromInt(1024 * 1024)).BigInt().Uint64(),
			UserID:   userID.(uint),
			Status:   models.OrderWait,
		}
		item.Price, _ = decimal.NewFromString(req.Amount)
		if len(item.Hash) > 0 {
			item.Status = models.OrderPending
		}
		if err := db.Save(item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

// @Summary add traffic bill
// @Schemes
// @Description add traffic bill
// @Tags bills
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/bills/traffic [post]
func AddBillsTrafficHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req BillReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		userID, _ := c.Get("id")
		item := &models.Order{
			PType:    models.ProductType_Traffic,
			Hash:     req.Hash,
			Quantity: req.Size.Mul(decimal.NewFromInt(1024 * 1024)).BigInt().Uint64(),
			UserID:   userID.(uint),
			Status:   models.OrderWait,
		}
		item.Price, _ = decimal.NewFromString(req.Amount)
		if len(item.Hash) > 0 {
			item.Status = models.OrderPending
		}
		if err := db.Save(item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

// @Summary traffic price
// @Schemes
// @Description traffic price
// @Tags bills
// @Accept json
// @Produce json
// @Param   size     query    int     true        "buy size"
// @Success 200 string ok
// @Router /api/v1/buy/traffic [get]
func BuyTrafficHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var item models.Product
		if err := db.Preload("Currency").Where("p_type = ?", models.ProductType_Storage).Find(&item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, &map[string]interface{}{
			"size":      item.Quantity,
			"amount":    item.Price,
			"receiptor": item.Currency.Receiptor,
		}))
	}
}

// @Summary traffic price
// @Schemes
// @Description traffic price
// @Tags bills
// @Accept json
// @Produce json
// @Param   size     query    string     true        "buy size"
// @Success 200 string ok
// @Router /api/v1/buy/storage [get]
func BuyStorageHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var item models.Product
		if err := db.Preload("Currency").Where("p_type = ?", models.ProductType_Storage).Find(&item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, &map[string]interface{}{
			"size":      item.Quantity,
			"amount":    item.Price,
			"receiptor": item.Currency.Receiptor,
		}))
	}
}
