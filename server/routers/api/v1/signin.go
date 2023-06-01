package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// @Summary Get a single signIn
// @Tags signIn
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} Response
// @Router /api/v1/signIns/{id} [get]
func GetSignIn(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var item models.SignIn
		res := db.Model(&models.SignIn{}).Where("id = ?", id).Find(&item)
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

// @Summary Get multiple signIns
// @Tags signIn
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/signIns [get]
func GetSignIns(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.SignIn{}).Order("id desc").Count(&total).Offset(int(offset)).Limit(int(pageSize))

		var items []*models.SignIn
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

type EditSignInReq struct {
	Quantity uint64 `json:"quantity"`
	Period   uint   `json:"period"`
}

// @Summary Update signIn
// @Tags signIn
// @Produce  json
// @Param id path int true "id"
// @Param data body EditSignInReq true "data"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/signIns/{id} [put]
func EditSignIn(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid id"))
			return
		}
		var req EditSignInReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}
		if req.Quantity == 0 {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid quantity"))
			return
		}
		period := models.SignInPeriod(req.Period)
		if period < models.SignInPeriod_End {
			c.JSON(OKCode, NewResponse(c, RequestCode, "invalid period"))
			return
		}

		res := db.Model(&models.SignIn{}).Where("id = ?", id).Updates(&models.SignIn{Quantity: req.Quantity, Period: period})
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
	}
}

// @Summary Get signIn Switch
// @Tags signIn
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/user/signIn [get]
func GetSignInSwitch(db *gorm.DB) func(c *gin.Context) {
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

		var items []*models.SignIn
		if err := db.Model(&models.SignIn{}).Where("enable = true").Find(&items).Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		signIn := false
		if !item.SignedIn.IsZero() {
			for _, i := range items {
				if models.SignInPeriod_Day == i.Period {
					signIn = time.Now().Day()-item.SignedIn.Day() > 1
				} else if models.SignInPeriod_Week == i.Period {
					_, w := time.Now().ISOWeek()
					_, w1 := item.SignedIn.ISOWeek()
					signIn = w-w1 > 1
				} else if models.SignInPeriod_Month == i.Period {
					signIn = time.Now().Month()-item.SignedIn.Month() > 1
				} else if models.SignInPeriod_Year == i.Period {
					signIn = time.Now().Year()-item.SignedIn.Year() > 1
				}
				if signIn {
					break
				}
			}
		}
		if len(items) == 0 {
			c.JSON(OKCode, NewResponse(c, OKCode, 0))
		} else if signIn {
			c.JSON(OKCode, NewResponse(c, OKCode, 1))
		} else {
			c.JSON(OKCode, NewResponse(c, OKCode, 2))
		}
	}
}

// @Summary Set signIn switch
// @Tags signIn
// @Produce  json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/signIns/switch [put]
func SetSignInSwitch(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var count int64
		if err := db.Model(&models.SignIn{}).Where("enable = true").Count(&count).Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		on_off := count > 0
		res := db.Model(&models.SignIn{}).Where("1 = 1").Update("enable", !on_off)
		if err := res.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, res.RowsAffected > 0))
	}
}
