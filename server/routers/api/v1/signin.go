package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"
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
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var item models.SignIn
		res := db.Model(&models.SignIn{}).Where("id = ?", id).Find(&item)
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

type EditSignInReq struct {
	Quantity uint64              `json:"quantity"`
	Period   models.SignInPeriod `json:"period"`
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
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var req EditSignInReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}
		if req.Quantity == 0 {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid quantity"))
			return
		}
		if req.Period < models.SignInPeriod_End {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid period"))
			return
		}

		res := db.Model(&models.SignIn{}).Where("id = ?", id).Updates(&models.SignIn{Quantity: req.Quantity, Period: req.Period})
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, res.RowsAffected > 0))
	}
}

// @Summary Get signIn Switch
// @Tags signIn
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/signIns/switch [get]
func GetSignInSwitch(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var count int64
		if err := db.Model(&models.SignIn{}).Where("enable = true").Count(&count).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, count > 0))
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
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		on_off := count > 0
		res := db.Model(&models.SignIn{}).Debug().Where("1 = 1").Update("enable", !on_off)
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		fmt.Println(res.RowsAffected)
		c.JSON(http.StatusOK, NewResponse(OKCode, res.RowsAffected > 0))
	}
}
