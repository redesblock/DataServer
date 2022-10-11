package routers

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	TokenExpireDuration = time.Hour * 8
	HeaderTokenKey      = "Authorization"
)

var MySecret = []byte("JWT SECRET")

type UserInfo struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

type MyClaims struct {
	UserInfo
	jwt.StandardClaims
}

func secret() []byte {
	if secret, ok := os.LookupEnv("DATA_SERVER_JWT_SECRET"); ok {
		return []byte(secret)
	}
	return MySecret
}

func GenToken(user UserInfo) (string, error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		user, // 自定义字段
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(), // 过期时间
			Issuer:    "ccc",                                      // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(secret())
}

func ParseToken(tokenString string) (*MyClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return secret(), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get(HeaderTokenKey)
		if authHeader == "" {
			log.Info("Authorization empty")
			c.JSON(http.StatusUnauthorized, NewResponse(AuthCode, "Empty Bearer Authorization"))
			c.Abort()
			return
		}
		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			log.Info("Invalid Authorization: ", authHeader)
			c.JSON(http.StatusUnauthorized, NewResponse(AuthCode, "Invalid Bearer Authorization"))
			c.Abort()
			return
		}
		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		mc, err := ParseToken(parts[1])
		if err != nil {
			log.Info("invalid Bearer token: ", authHeader)
			c.JSON(http.StatusUnauthorized, NewResponse(AuthCode, "Invalid Bearer Authorization"))
			c.Abort()
			return
		}
		// 将当前请求的username信息保存到请求的上下文c上
		c.Set("id", mc.ID)
		c.Set("email", mc.Email)

		// 更新过期时间
		if token, err := GenToken(mc.UserInfo); err == nil {
			c.Header(HeaderTokenKey, "Bearer "+token)
		}
		c.Next() // 后续的处理函数可以用过c.Get("user")来获取当前请求的用户信
	}
}
