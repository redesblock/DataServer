package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"
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
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var item models.User
		res := db.Model(&models.User{}).Where("id = ?", id).Find(&item)
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
		tx := db.Model(&models.User{}).Order("id desc").Not("role = ?", models.UserRole_User).Count(&total).Offset(int(offset)).Limit(int(pageSize))

		var items []models.User
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
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		item := &models.User{
			Email:    req.Email,
			Password: req.Password,
			Role:     models.UserRole_Oper,
		}
		res := db.Model(&models.User{}).Save(item)
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
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
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		var req EditUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		item := &models.User{
			Status: req.Status,
		}
		if len(req.Password) > 0 {
			item.Password = req.Password
		}
		res := db.Model(&models.User{}).Where("id = ?", id).Updates(item)
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, res.RowsAffected > 0))
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
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}
		res := db.Unscoped().Where("id = ?", id).Delete(&models.User{})
		if err := res.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, res.RowsAffected > 0))
	}
}
