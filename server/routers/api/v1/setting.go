package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
)

// @Summary user info
// @Schemes
// @Description user info
// @Security ApiKeyAuth
// @Tags setting
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/user [get]
func GetUserHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")
		var item models.User
		ret := db.Model(&models.User{}).Where("id = ?", userID).Find(&item)
		if err := ret.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		if ret.RowsAffected == 0 {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, nil))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, item))
	}
}

type UserReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}

// @Summary update user info
// @Schemes
// @Description update user info
// @Security ApiKeyAuth
// @Tags setting
// @Accept json
// @Produce json
// @Param user body UserReq true "user setting"
// @Success 200 {object} Response
// @Router /api/v1/user [post]
func AddUserHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req UserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(OKCode, NewResponse(c, RequestCode, err.Error()))
			return
		}
		req.Password = Sha256(req.Password)

		userID, _ := c.Get("id")

		var item models.User
		ret := db.Model(&models.User{}).Where("id = ?", userID).Find(&item)
		if err := ret.Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		if ret.RowsAffected == 0 {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, "not found"))
			return
		}

		item.FirstName = req.FirstName
		item.LastName = req.LastName
		item.Email = req.Email
		if len(req.NewPassword) > 0 {
			if item.Password != req.Password {
				c.JSON(OKCode, NewResponse(c, ExecuteCode, "wrong password"))
				return
			}
			req.NewPassword = Sha256(req.NewPassword)
			item.Password = req.NewPassword
		}

		if err := db.Save(item).Error; err != nil {
			c.JSON(OKCode, NewResponse(c, ExecuteCode, err))
			return
		}
		c.JSON(OKCode, NewResponse(c, OKCode, item))
	}
}

// @Summary list user actions
// @Schemes
// @Description pagination query user actions
// @Security ApiKeyAuth
// @Tags dashboard
// @Accept json
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} Response
// @Router /api/v1/user/actions [get]
func UserActionsHandler(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")

		var total int64
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize
		tx := db.Model(&models.UserAction{}).Order("id desc").Where("user_id = ?", userID).Count(&total).Offset(int(offset)).Limit(int(pageSize))

		var items []*models.UserAction
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
