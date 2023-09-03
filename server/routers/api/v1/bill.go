package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"github.com/redesblock/dataserver/server/pay"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"strings"
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

type BillReq struct {
	Size           decimal.Decimal       `json:"quantity"`
	Price          decimal.Decimal       `json:"price"`
	SpecialProduct uint                  `json:"special_product"`
	PaymentChannel models.PaymentChannel `json:"payment_channel"`
	Currency       uint                  `json:"currency"`
	Coupon         uint                  `json:"coupon"`
	Description    string                `json:"desc"`
	Hash           string                `json:"hash"`
}

func (r *BillReq) convertToOrder(db *gorm.DB, p_type models.ProductType) (quantity uint64, price decimal.Decimal, discount decimal.Decimal, err error) {
	var item models.Product
	ret := db.Where("p_type = ?", p_type).Find(&item)
	if err = ret.Error; err != nil {
		return
	} else if ret.RowsAffected == 0 {
		err = fmt.Errorf("invalid p_type")
		return
	}

	var citem models.Currency
	ret = db.Find(&citem, r.Currency)
	if err = ret.Error; err != nil {
		return
	}
	if ret.RowsAffected == 0 {
		err = fmt.Errorf("invalid currency")
		return
	}

	if r.SpecialProduct != 0 {
		var item2 models.SpecialProduct
		ret := db.Where("reserve > 0").Find(&item2, r.SpecialProduct)
		if err = ret.Error; err != nil {
			return
		}
		if ret.RowsAffected == 0 {
			err = fmt.Errorf("invalid special_product")
			return
		}
		quantity = item2.Quantity
		price = decimal.NewFromInt(int64(quantity)).Div(decimal.NewFromInt(int64(item.Quantity))).Mul(item.Price)
		discount = item2.Discount.Mul(price).Div(decimal.NewFromInt(10))
	} else {
		quantity = r.Size.BigInt().Uint64()
		if quantity == 0 {
			discount = r.Price
			if r.Coupon > 0 {
				var item2 models.UserCoupon
				ret := db.Where("status = ?", models.UserCouponStatus_Normal).Preload("Coupon").Find(&item2, r.Coupon)
				if err = ret.Error; err != nil {
					return
				}
				if ret.RowsAffected == 0 {
					err = fmt.Errorf("invalid coupon")
					return
				}
				price = discount.Mul(decimal.NewFromInt(10)).Div(item2.Coupon.Discount)
			}
			if strings.ToUpper(citem.Symbol) != "MOP" && strings.ToUpper(citem.Symbol) != "USDT" {
				price, _ = decimal.NewFromString(price.StringFixed(2))
				discount, _ = decimal.NewFromString(discount.StringFixed(2))
			}
			quantity = price.Div(citem.Rate).Div(item.Price).Mul(decimal.NewFromInt(int64(item.Quantity))).BigInt().Uint64()
		} else {
			price = decimal.NewFromInt(int64(quantity)).Div(decimal.NewFromInt(int64(item.Quantity))).Mul(item.Price)
			discount = price
			if r.Coupon > 0 {
				var item2 models.UserCoupon
				ret := db.Where("status = ?", models.UserCouponStatus_Normal).Preload("Coupon").Find(&item2, r.Coupon)
				if err = ret.Error; err != nil {
					return
				}
				if ret.RowsAffected == 0 {
					err = fmt.Errorf("invalid coupon")
					return
				}
				discount = item2.Coupon.Discount.Mul(price).Div(decimal.NewFromInt(10))
			}
			price = price.Mul(citem.Rate)
			discount = discount.Mul(citem.Rate)
			if strings.ToUpper(citem.Symbol) != "MOP" && strings.ToUpper(citem.Symbol) != "USDT" {
				price, _ = decimal.NewFromString(price.StringFixed(2))
				discount, _ = decimal.NewFromString(discount.StringFixed(2))
			}
		}
	}

	return
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
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		if len(req.Description) == 0 {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid desc"))
			return
		}

		//var citem models.Currency
		//ret := db.Find(&citem, req.Currency)
		//if err := ret.Error; err != nil {
		//	c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
		//	return
		//}
		//if ret.RowsAffected == 0 {
		//	c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid currency")))
		//	return
		//}
		//if int64(citem.Payment)&int64(req.PaymentChannel) == 0 {
		//	c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid payment channel")))
		//	return
		//}

		var uitem models.UserCoupon
		if req.Coupon > 0 {
			ret := db.Preload("Coupon").Where("status = ?", models.UserCouponStatus_Normal).Find(&uitem, req.Coupon)
			if err := ret.Error; err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
				return
			}
			if ret.RowsAffected == 0 {
				c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid coupon")))
				return
			}
			if int64(uitem.PType)&int64(models.ProductType_Storage) == 0 {
				c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid coupon")))
				return
			}
		}

		quantity, price, discount, err := req.convertToOrder(db, models.ProductType_Storage)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err))
			return
		}

		userID, _ := c.Get("id")
		item := &models.Order{
			OrderID:      generateOrderID(),
			PType:        models.ProductType_Storage,
			Payment:      req.PaymentChannel,
			Quantity:     quantity,
			UserID:       userID.(uint),
			Status:       models.OrderWait,
			Hash:         req.Hash,
			Price:        price,
			Discount:     discount,
			CurrencyID:   req.Currency,
			PaymentTime:  models.UnlimitedTime,
			UserCouponID: uitem.ID,
			Discount1:    uitem.Coupon.Discount,
		}
		if req.Coupon == 0 {
			item.Discount1 = decimal.NewFromInt(10)
		}

		if req.PaymentChannel == models.PaymentChannel_Crypto {
			if len(item.Hash) > 0 {
				item.Status = models.OrderPending
			} else {
				c.JSON(OKCode, NewResponse(c, ExecuteCode, "no crypto hash"))
				return
			}
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if req.SpecialProduct != 0 {
				var item2 models.SpecialProduct
				ret := tx.Where("reserve > 0").Find(&item2, req.SpecialProduct)
				if err = ret.Error; err != nil {
					return err
				}
				if ret.RowsAffected == 0 {
					err = fmt.Errorf("invalid special_product")
					return err
				}
				item2.Sold++
				item2.Reserve--
				if err := tx.Save(item2).Error; err != nil {
					return err
				}
			}
			if err := tx.Save(item).Error; err != nil {
				return err
			}
			if req.PaymentChannel == models.PaymentChannel_Alipay {
				res, err := pay.AliPayTrade(req.Description, item.OrderID, item.Discount.String())
				if err != nil {
					return err
				}
				c.JSON(OKCode, NewResponse(c, OKCode, map[string]interface{}{"order_id": item.ID, "url": res}))
				return nil
				//c.Redirect(http.StatusTemporaryRedirect, res)
			} else if req.PaymentChannel == models.PaymentChannel_WeChat {
				res, err := pay.WXTrade(req.Description, item.OrderID, item.Discount.String())
				if err != nil {
					return err
				}
				c.JSON(OKCode, NewResponse(c, OKCode, map[string]interface{}{"order_id": item.ID, "url": res}))
				//c.Redirect(http.StatusTemporaryRedirect, res)
				return nil
			} else if req.PaymentChannel == models.PaymentChannel_Stripe {
				currency := "usd"
				if req.Currency == 3 {
					currency = "cny"
				}
				res, err := pay.StripeTrade(req.Description, item.OrderID, item.Discount.String(), currency)
				if err != nil {
					return err
				}
				//c.JSON(OKCode, NewResponse(c, OKCode, map[string]interface{}{"order_id": item.ID, "url": res}))
				c.JSON(OKCode, NewResponse(c, OKCode, res))
				//c.Redirect(http.StatusTemporaryRedirect, res)
				return nil
			} else if req.PaymentChannel == models.PaymentChannel_Crypto {
				c.JSON(OKCode, NewResponse(c, OKCode, map[string]interface{}{"order_id": item.ID, "url": viper.GetString("bsc.browser") + "tx/" + req.Hash}))
				return nil
			} else {
				return fmt.Errorf("not support payment channel")
			}
			return nil
		}); err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
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
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		if len(req.Description) == 0 {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid desc"))
			return
		}

		//var citem models.Currency
		//ret := db.Find(&citem, req.Currency)
		//if err := ret.Error; err != nil {
		//	c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
		//	return
		//}
		//if ret.RowsAffected == 0 {
		//	c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid currency")))
		//	return
		//}
		//if int64(citem.Payment)&int64(req.PaymentChannel) == 0 {
		//	c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid payment channel")))
		//	return
		//}

		var uitem models.UserCoupon
		if req.Coupon > 0 {
			ret := db.Preload("Coupon").Where("status = ?", models.UserCouponStatus_Normal).Find(&uitem, req.Coupon)
			if err := ret.Error; err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
				return
			}
			if ret.RowsAffected == 0 {
				c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid coupon")))
				return
			}
			if int64(uitem.PType)&int64(models.ProductType_Storage) == 0 {
				c.JSON(OKCode, NewResponse(c, RequestCode, fmt.Errorf("invalid coupon")))
				return
			}
		}

		quantity, price, discount, err := req.convertToOrder(db, models.ProductType_Storage)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err))
			return
		}

		userID, _ := c.Get("id")
		item := &models.Order{
			OrderID:      generateOrderID(),
			PType:        models.ProductType_Traffic,
			Payment:      req.PaymentChannel,
			Quantity:     quantity,
			UserID:       userID.(uint),
			Status:       models.OrderWait,
			Hash:         req.Hash,
			Price:        price,
			Discount:     discount,
			CurrencyID:   req.Currency,
			PaymentTime:  models.UnlimitedTime,
			UserCouponID: uitem.ID,
			Discount1:    uitem.Coupon.Discount,
		}
		if req.Coupon == 0 {
			item.Discount1 = decimal.NewFromInt(10)
		}

		if req.PaymentChannel == models.PaymentChannel_Crypto {
			if len(item.Hash) > 0 {
				item.Status = models.OrderPending
			} else {
				c.JSON(OKCode, NewResponse(c, ExecuteCode, "no crypto hash"))
				return
			}
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if req.SpecialProduct != 0 {
				var item2 models.SpecialProduct
				ret := tx.Where("reserve > 0").Find(&item2, req.SpecialProduct)
				if err = ret.Error; err != nil {
					return err
				}
				if ret.RowsAffected == 0 {
					err = fmt.Errorf("invalid special_product")
					return err
				}
				item2.Sold++
				item2.Reserve--
				if err := tx.Save(item2).Error; err != nil {
					return err
				}
			}
			if err := tx.Save(item).Error; err != nil {
				return err
			}
			if req.PaymentChannel == models.PaymentChannel_Alipay {
				res, err := pay.AliPayTrade(req.Description, item.OrderID, item.Discount.String())
				if err != nil {
					return err
				}
				c.JSON(OKCode, NewResponse(c, OKCode, map[string]interface{}{"order_id": item.ID, "url": res}))
				//c.Redirect(http.StatusTemporaryRedirect, res)
				return nil
			} else if req.PaymentChannel == models.PaymentChannel_WeChat {
				res, err := pay.WXTrade(req.Description, item.OrderID, item.Discount.String())
				if err != nil {
					return err
				}
				c.JSON(OKCode, NewResponse(c, OKCode, map[string]interface{}{"order_id": item.ID, "url": res}))
				//c.Redirect(http.StatusTemporaryRedirect, res)
				return nil
			} else if req.PaymentChannel == models.PaymentChannel_Stripe {
				currency := "usd"
				if req.Currency == 3 {
					currency = "cny"
				}
				res, err := pay.StripeTrade(req.Description, item.OrderID, item.Discount.String(), currency)
				if err != nil {
					return err
				}
				c.JSON(OKCode, NewResponse(c, OKCode, res))
				//c.Redirect(http.StatusTemporaryRedirect, res)
				return nil
			} else if req.PaymentChannel == models.PaymentChannel_Crypto {
				c.JSON(OKCode, NewResponse(c, OKCode, map[string]interface{}{"order_id": item.ID, "url": viper.GetString("bsc.browser") + "tx/" + req.Hash}))
				return nil
			} else {
				return fmt.Errorf("not support payment channel")
			}
			return nil
		}); err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
	}
}

// @Summary traffic price
// @Schemes
// @Description traffic price
// @Tags bills
// @Accept json
// @Produce json
// @Param data body BillReq true "data"
// @Success 200 string ok
// @Router /api/v1/buy/traffic [get]
func BuyTrafficHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req BillReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		quantity, price, discount, err := req.convertToOrder(db, models.ProductType_Traffic)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err))
			return
		}
		fmt.Println(quantity, price, discount, req.Size, req.Price, req.Currency)
		c.JSON(OKCode, NewResponse(c, OKCode, &map[string]interface{}{
			"quantity": quantity,
			"price":    price,
			"discount": discount,
		}))
	}
}

// @Summary traffic price
// @Schemes
// @Description traffic price
// @Tags bills
// @Accept json
// @Produce json
// @Param data body BillReq true "data"
// @Success 200 string ok
// @Router /api/v1/buy/storage [get]
func BuyStorageHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req BillReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		quantity, price, discount, err := req.convertToOrder(db, models.ProductType_Storage)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err))
			return
		}
		fmt.Println(quantity, price, discount, req.Size, req.Price, req.Currency)
		c.JSON(OKCode, NewResponse(c, OKCode, &map[string]interface{}{
			"quantity": quantity,
			"price":    price,
			"discount": discount,
		}))
	}
}
