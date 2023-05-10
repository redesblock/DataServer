package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/models"
	"gorm.io/gorm"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var codeStore = captcha.NewMemoryStore(10000, 10*time.Minute)

// @Summary user login
// @Tags login
// @Accept json
// @Produce json
// @Param user body LoginReq true "user info"
// @Success 200 string token
// @Router /api/v1/login [post]
func Login(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req LoginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}
		if !VerifyEmailFormat(req.Email) {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid email"))
			return
		}
		if len(req.Password) == 0 {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid password"))
			return
		}

		var item models.User
		ret := db.Model(&models.User{}).Where("email = ?", req.Email).Find(&item)
		if err := ret.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		req.Password = Sha256(req.Password)

		if ret.RowsAffected == 0 {
			item = models.User{
				Email:    req.Email,
				Password: req.Password,
				Role:     models.UserRole_User,
			}
			if err := db.Save(&item).Error; err != nil {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
				return
			}
		}
		if item.Password != req.Password {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, "wrong password"))
			return
		}

		token, err := GenToken(UserInfo{
			ID:    item.ID,
			Email: item.Email,
			Role:  item.Role,
		})
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		db.Save(&models.UserAction{
			ActionType: models.UserActionType_Login,
			Email:      item.Email,
			IP:         RemoteIp(c.Request),
			UserID:     item.ID,
		})
		c.Header(HeaderTokenKey, "Bearer "+token)
		c.JSON(http.StatusOK, NewResponse(OKCode, "Bearer "+token))
	}
}

type ForgotReq struct {
	Email string `json:"email"`
}

// @Summary user forgot password
// @Tags login
// @Accept json
// @Produce json
// @Param user body ForgotReq true "user info"
// @Success 200 string token
// @Router /api/v1/forgot [post]
func Forgot(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req ForgotReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}
		if !VerifyEmailFormat(req.Email) {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid email"))
			return
		}

		var item models.User
		ret := db.Model(&models.User{}).Where("email = ?", req.Email).Find(&item)
		if err := ret.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		if ret.RowsAffected == 0 {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "email not exist"))
			return
		}

		bts := captcha.RandomDigits(captcha.DefaultLen)
		codeStore.Set(req.Email, bts)
		var code string
		for _, num := range bts {
			code += strconv.Itoa(int(num))
		}

		if err := SendGoMail([]string{req.Email}, "Reset password", fmt.Sprintf(EmailContentTemplate_RESET, code)); err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, nil))
	}
}

type ResetReq struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

// @Summary user reset password
// @Tags login
// @Accept json
// @Produce json
// @Param user body ResetReq true "user info"
// @Success 200 string token
// @Router /api/v1/reset [post]
func Reset(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req ResetReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}
		if !VerifyEmailFormat(req.Email) {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid email"))
			return
		}

		var item models.User
		ret := db.Model(&models.User{}).Where("email = ?", req.Email).Find(&item)
		if err := ret.Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		bts := codeStore.Get(req.Email, true)
		var code string
		for _, num := range bts {
			code += strconv.Itoa(int(num))
		}

		if len(req.Code) == 0 || code != req.Code {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, "wrong code"))
			return
		}

		item.Password = Sha256(req.Password)
		if err := db.Save(&item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		db.Save(&models.UserAction{
			ActionType: models.UserActionType_Forgot,
			Email:      item.Email,
			IP:         RemoteIp(c.Request),
			UserID:     item.ID,
		})
		c.JSON(http.StatusOK, NewResponse(OKCode, nil))
	}
}

func VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func RemoteIp(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get("X-Forwarded-For"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

func Sha256(src string) string {
	m := sha256.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}
