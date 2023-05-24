package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// @Summary Get a single coupon
// @Tags coupon
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} Response
// @Router /api/v1/coupons/{id} [get]
func GetCoupon(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var item models.Coupon
		res := db.Model(&models.Coupon{}).Where("id = ?", id).Find(&item)
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

// @Summary Get multiple coupons
// @Tags coupon
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/coupons [get]
func GetCoupons(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.Coupon{}).Order("id desc").Count(&total).Offset(int(offset)).Limit(int(pageSize))

		var items []models.Coupon
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

type AddCouponReq struct {
	Name               string             `json:"name" gorm:"unique"`
	CouponType         models.CouponType  `json:"coupon_type"`
	PType              models.ProductType `json:"product_type"`
	Discount           decimal.Decimal    `json:"discount"`
	StorageQuantityMin uint64             `json:"storage_quantity_min"`
	StorageQuantityMax uint64             `json:"storage_quantity_max"`
	TrafficQuantityMin uint64             `json:"traffic_quantity_min"`
	TrafficQuantityMax uint64             `json:"traffic_quantity_max"`
	StartTime          int64              `json:"start_time"`
	EndTime            int64              `json:"end_time"`
	Reserve            uint64             `json:"reserve"`
	MaxClaim           uint64             `json:"max_claim"`
}

// @Summary Add coupon
// @Tags coupon
// @Produce  json
// @Param data body AddCouponReq true "data"
// @Success 200 {object} Response
// @Router /api/v1/coupons [post]
func AddCoupon(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req AddCouponReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		item := &models.Coupon{
			Name:               req.Name,
			CouponType:         req.CouponType,
			PType:              req.PType,
			Discount:           req.Discount,
			Reserve:            req.Reserve,
			MaxClaim:           req.MaxClaim,
			StorageQuantityMin: req.StorageQuantityMin,
			StorageQuantityMax: req.StorageQuantityMax,
			TrafficQuantityMin: req.TrafficQuantityMin,
			TrafficQuantityMax: req.TrafficQuantityMax,
			StartTime:          time.Unix(req.StartTime, 0),
			EndTime:            time.Unix(req.EndTime, 0),
			Status:             models.CouponStatus_NotStart,
		}
		res := db.Model(&models.Coupon{}).Save(item)
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, item))
	}
}

type EditCouponReq struct {
	Reserve  uint64 `json:"reserve"`
	MaxClaim uint64 `json:"max_claim"`
}

// @Summary Update coupon
// @Tags coupon
// @Produce  json
// @Param id path int true "id"
// @Param data body EditCouponReq true "data"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/coupons/{id} [put]
func EditCoupon(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var req EditCouponReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		res := db.Model(&models.Coupon{}).Where("id = ?", id).Updates(&models.Coupon{
			Reserve:  req.Reserve,
			MaxClaim: req.MaxClaim,
		})
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
	}
}

// @Summary Delete article
// @Tags coupon
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/coupons/{id} [delete]
func DeleteCoupon(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		res := db.Unscoped().Where("id = ?", id).Delete(&models.Coupon{})
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
	}

}
