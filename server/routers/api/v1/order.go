package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"strconv"
	"time"
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
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var item models.Order
		res := db.Model(&models.Order{}).Where("id = ?", id).Preload("User").Preload("Currency").Find(&item)
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

// @Summary Get multiple orders
// @Tags order
// @Produce json
// @Param   p_type     query    int     false        "folder id"
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Param   start   query    string     true        "start"
// @Param   end   query    string     true        "end"
// @Param   order_id   query    string     true        "order id"
// @Param   payment   query    string     true        "payment"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/orders [get]
func GetOrders(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.Order{}).Order("id desc")
		if len(c.Query("p_type")) > 0 {
			pType, err := strconv.ParseInt(c.Query("p_type"), 10, 64)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, err))
				return
			}
			tx = tx.Where("p_type = ?", pType)
		}
		start := c.Query("start")
		end := c.Query("end")
		if len(start) > 0 && len(end) > 0 {
			startTime, err := time.Parse("2006-01-02", start)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
				return
			}
			endTime, err := time.Parse("2006-01-02", end)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
				return
			}
			if startTime.After(endTime) {
				tx = tx.Where("created_at BETWEEN ? AND ?", endTime, startTime)
			} else {
				tx = tx.Where("created_at BETWEEN ? AND ?", startTime, endTime)
			}
		}
		if order := c.Query("order_id"); len(order) > 0 {
			tx = tx.Where("order_id = ?", order)
		}
		if payment := c.Query("payment"); len(payment) > 0 {
			tx = tx.Where("payment_account = ?", payment)
		}

		var items []models.Order
		if err := tx.Count(&total).Offset(int(offset)).Limit(int(pageSize)).Preload("User").Preload("Currency").Find(&items).Error; err != nil {
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

//type AddOrderReq struct {
//}
//
//// @Summary Add order
//// @Produce  json
//// @Param data body AddOrderReq true "data"
//// @Success 200 {object} Response
//// @Failure 500 {object} Response
//// @Router /api/v1/articles [post]
//func AddOrder(db *gorm.DB) func(c *gin.Context) {
//	return func(c *gin.Context) {
//		var req AddOrderReq
//		if err := c.ShouldBindJSON(&req); err != nil {
//			c.JSON(OKCode, NewResponse(c,RequestCode, err.Error()))
//			return
//		}
//
//		item := &models.Order{}
//		res := db.Model(&models.Node{}).Save(item)
//		if err := res.Error; err != nil {
//			c.JSON(OKCode, NewResponse(c,ExecuteCode, err))
//			return
//		}
//		c.JSON(OKCode, NewResponse(c,OKCode, item))
//	}
//}
//
//type EditOrderReq struct {
//}
//
//// @Summary Update order
//// @Produce  json
//// @Param id path int true "id"
//// @Param data body EditOrderReq true "data"
//// @Success 200 {object} Response
//// @Failure 500 {object} Response
//// @Router /api/v1/orders/{id} [put]
//func EditOrder(db *gorm.DB) func(c *gin.Context) {
//	return func(c *gin.Context) {
//		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
//		if err != nil {
//			c.JSON(OKCode, NewResponse(c,RequestCode, "invalid id"))
//			return
//		}
//		var req EditOrderReq
//		if err := c.ShouldBindJSON(&req); err != nil {
//			c.JSON(OKCode, NewResponse(c,RequestCode, err.Error()))
//			return
//		}
//
//		res := db.Model(&models.Node{}).Where("id = ?", id).Updates(&models.Order{})
//		if err := res.Error; err != nil {
//			c.JSON(OKCode, NewResponse(c,ExecuteCode, err))
//			return
//		}
//		c.JSON(OKCode, NewResponse(c,OKCode, res.RowsAffected > 0))
//	}
//}
