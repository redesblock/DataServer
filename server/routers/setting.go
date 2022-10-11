package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"net/http"
)

// @Summary user info
// @Schemes
// @Description user info
// @Security ApiKeyAuth
// @Tags setting
// @Accept json
// @Produce json
// @Success 200 {object} dataservice.User
// @Router /user [get]
func GetUserHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")
		item, err := db.FindUserByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
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
// @Success 200 {object} dataservice.User
// @Router /user [post]
func AddUserHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req UserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}
		req.Password = Sha256(req.Password)

		userID, _ := c.Get("id")
		item, err := db.FindUserByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		item.FirstName = req.FirstName
		item.LastName = req.LastName
		item.Email = req.Email
		if len(req.NewPassword) > 0 {
			if item.Password != req.Password {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, "wrong password"))
				return
			}
			req.NewPassword = Sha256(req.NewPassword)
			item.Password = req.NewPassword
		}

		if err := db.Save(item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}
