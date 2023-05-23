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

// @Summary Get a single user
// @Tags user
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} Response
// @Router /api/v1/users/{id} [get]
func GetUser(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var item models.User
		res := db.Model(&models.User{}).Where("id = ?", id).Find(&item)
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

// @Summary Get multiple users
// @Tags user
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/users [get]
func GetUsers(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.User{}).Order("id desc").Where("role = ?", models.UserRole_User).Count(&total).Offset(int(offset)).Limit(int(pageSize))

		var items []models.User
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

// @Summary Get multiple users
// @Tags user
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/operators [get]
func GetOperators(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.User{}).Order("id desc").Where("role <> ?", models.UserRole_User).Count(&total).Offset(int(offset)).Limit(int(pageSize))

		var items []models.User
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

type AddUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// @Summary Add user
// @Tags user
// @Produce  json
// @Param data body AddUserReq true "data"
// @Success 200 {object} Response
// @Router /api/v1/users [post]
func AddUser(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req AddUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}
		item := &models.User{
			Email: req.Email,
		}
		if res := db.Model(&models.User{}).Where("email = ?", req.Email).Find(item); res.Error != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, res.Error))
			return
		} else if res.RowsAffected > 0 {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, fmt.Errorf("user alreay exist")))
			return
		}
		item.Email = req.Email
		item.Password = Sha256(req.Password)
		item.Role = models.UserRole_Oper
		res := db.Model(&models.User{}).Save(item)
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, item))
	}
}

type EditUserReq struct {
	Password string            `json:"password"`
	Status   models.UserStatus `json:"status"`
}

// @Summary Update user
// @Tags user
// @Produce  json
// @Param id path int true "id"
// @Param data body EditUserReq true "data"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/users/{id} [put]
func EditUser(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var req EditUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}

		item := &models.User{
			Status: req.Status,
		}
		if len(req.Password) > 0 {
			item.Password = Sha256(req.Password)
			res := db.Model(&models.User{}).Where("id = ?", id).Updates(item)
			if err := res.Error; err != nil {
				c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
				return
			}
			c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
		} else {
			res := db.Model(&models.User{}).Where("id = ?", id).Update("status", item.Status)
			if err := res.Error; err != nil {
				c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
				return
			}
			c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
		}
	}
}

// @Summary Delete article
// @Tags user
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/users/{id} [delete]
func DeleteUser(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		res := db.Unscoped().Where("id = ?", id).Delete(&models.User{})
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
	}
}

// @Summary user signed in
// @Tags signIn
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Param   p_type    query    int     false        "type"
// @Success 200 {object} Response
// @Router /api/v1/claimed [get]
func GetClaimed(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize

		tx := db.Model(&models.UserCoupon{}).Order("id desc").Where("user_id = ?", userID)
		if usable := c.Query("usable"); len(usable) > 0 {
			if strings.ToLower(usable) == "true" {
				tx = tx.Where("used = ?", false)
			} else {
				tx = tx.Where("used = ?", true)
			}
		}
		if len(c.Query("p_type")) > 0 {
			pType, err := strconv.ParseInt(c.Query("p_type"), 10, 64)
			if err != nil {
				c.JSON(OKCode, NewResponse(c, RequestCode, err))
				return
			}
			tx = tx.Where("p_type in ?", []int64{int64(models.ProductType_Storage) & pType, int64(models.ProductType_Traffic) & pType})
		}

		var items []*models.UserCoupon
		ret := tx.Count(&total).Offset(int(offset)).Limit(int(pageNum)).Preload("Coupon").Find(&items)
		if err := ret.Error; err != nil {
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
			Items: func() []*models.Coupon {
				var ret []*models.Coupon
				for _, item := range items {
					ret = append(ret, &item.Coupon)
				}
				return ret
			}(),
		}))
	}
}

// @Summary user signed in
// @Tags signIn
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Router /api/v1/unclaimed [get]
func GetUnclaimed(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.Coupon{}).Order("id desc").Where("reserve > 0")

		var items []*models.Coupon
		ret := tx.Count(&total).Offset(int(offset)).Limit(int(pageNum)).Find(&items)
		if err := ret.Error; err != nil {
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

// @Summary user signed in
// @Tags signIn
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/unclaimed/{:id} [get]
func GetClaim(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var item models.Coupon
		ret := db.Model(&models.Coupon{}).Where("id = ?", id).Find(&item)
		if err := ret.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		if ret.RowsAffected == 0 {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, "not found"))
			return
		}

		if item.Reserve == 0 {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, "no reserve"))
			return
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if item.MaxClaim != 0 {
				var count int64
				if err := db.Model(&models.UserCoupon{}).Where("user_id = ? AND coupon_id = ?", userID, id).Count(&count).Error; err != nil {
					return err
				}
				if uint64(count) >= item.MaxClaim {
					return fmt.Errorf("over max claim")
				}
			}
			item.Reserve--
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
			if err := tx.Save(&models.UserCoupon{
				UserID:   userID.(uint),
				CouponID: uint(id),
				PType:    item.PType,
				Used:     false,
			}).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}

		c.JSON(OKCode, NewResponse(c, OKCode, item))
		return
	}
}

// @Summary user signed in
// @Tags signIn
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/user/signedIn [get]
func GetSignedIn(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")
		var item models.User
		ret := db.Model(&models.User{}).Where("id = ?", userID).Find(&item)
		if err := ret.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		if ret.RowsAffected == 0 {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, "user not found"))
			return
		}

		var signIns []*models.SignIn
		if err := db.Model(&models.SignIn{}).Where("enable = true").Find(&signIns).Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}

		var storage uint64
		var traffic uint64
		for _, signIn := range signIns {
			enable := false
			if models.SignInPeriod_Day == signIn.Period {
				enable = time.Now().Day()-item.SignedIn.Day() > 1
			} else if models.SignInPeriod_Week == signIn.Period {
				_, w := time.Now().ISOWeek()
				_, w1 := item.SignedIn.ISOWeek()
				enable = w-w1 > 1
			} else if models.SignInPeriod_Month == signIn.Period {
				enable = time.Now().Month()-item.SignedIn.Month() > 1
			} else if models.SignInPeriod_Year == signIn.Period {
				enable = time.Now().Year()-item.SignedIn.Year() > 1
			}
			if !enable {
				continue
			}

			switch signIn.PType {
			case models.ProductType_Storage:
				storage += signIn.Quantity
			case models.ProductType_Traffic:
				traffic += signIn.Quantity
			}
		}

		item.TotalStorage += storage
		item.TotalTraffic += traffic
		item.SignedIn = time.Now()
		if err := db.Transaction(func(tx *gorm.DB) error {
			err := tx.Debug().Save(&item).Error
			if err != nil {
				return err
			}
			if storage > 0 {
				err := tx.Save(&models.Order{
					OrderID:  generateOrderID(),
					PType:    models.ProductType_Storage,
					Quantity: storage,
					Payment:  models.PaymentChannel_SignIn,
					Status:   models.OrderComplete,
					UserID:   userID.(uint),
				}).Error
				if err != nil {
					return err
				}
			}
			if traffic > 0 {
				err := tx.Save(&models.Order{
					OrderID:  generateOrderID(),
					PType:    models.ProductType_Traffic,
					Quantity: traffic,
					Payment:  models.PaymentChannel_SignIn,
					Status:   models.OrderComplete,
					UserID:   userID.(uint),
				}).Error
				if err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, fmt.Sprintf("Congratulations on your successful sign-in! You have now received a free %s storage space and %s download traffic.", models.ByteSize(storage), models.ByteSize(traffic))))
	}
}
