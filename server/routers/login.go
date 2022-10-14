package routers

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"math"
	"net"
	"net/http"
	"regexp"
)

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// @Summary user login
// @Schemes
// @Description user login
// @Tags login
// @Accept json
// @Produce json
// @Param user body LoginReq true "user info"
// @Success 200 string token
// @Router /login [post]
func LoginHandler(db *dataservice.DataService) func(c *gin.Context) {
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

		item, err := db.FindUserByEmail(req.Email)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		req.Password = Sha256(req.Password)

		if item == nil {
			item = &dataservice.User{
				Email:        req.Email,
				Password:     req.Password,
				TotalStorage: math.MaxUint64,
				TotalTraffic: math.MaxUint64,
			}
			if err := db.Save(item).Error; err != nil {
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
		})
		if err != nil {
			panic(err)
		}

		db.Save(&dataservice.UserAction{
			Action: "Login",
			Email:  item.Email,
			IP:     RemoteIp(c.Request),
			UserID: item.ID,
		})
		c.Header(HeaderTokenKey, "Bearer "+token)
		c.JSON(http.StatusOK, NewResponse(OKCode, "Bearer "+token))
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
