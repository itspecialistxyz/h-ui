package service

import (
	"errors"
	"h-ui/dao"
	"h-ui/model/bo"
	"h-ui/model/constant"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

const TokenExpireDuration = time.Hour * 24

type MyClaims struct {
	AccountBo bo.AccountBo `json:"account"`
	jwt.StandardClaims
}

func GenToken(accountBo bo.AccountBo) (string, error) {
	c := MyClaims{
		AccountBo: accountBo,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(),
			Issuer:    "h-ui",
		},
	}

	config, err := dao.GetConfig("key = ?", constant.JwtSecret)
	if err != nil {
		logrus.Errorf("Failed to get JWT secret from database: %v", err)
		return "", errors.New(constant.SysError)
	}

	if config.Value == nil {
		logrus.Error("JWT secret is not configured in database")
		return "", errors.New("JWT secret is not set")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(*config.Value))
}

func ParseToken(tokenString string) (*MyClaims, error) {
	// Check for empty token string
	if tokenString == "" {
		return nil, errors.New(constant.UnauthorizedError)
	}

	config, err := dao.GetConfig("key = ?", constant.JwtSecret)
	if err != nil {
		logrus.Errorf("Failed to get JWT secret: %v", err)
		return nil, errors.New(constant.SysError)
	}
	if config.Value == nil {
		logrus.Error("JWT secret is not configured in database")
		return nil, errors.New("JWT secret is not set")
	}
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logrus.Errorf("Invalid JWT signing method: %v", token.Method)
			return nil, errors.New("invalid signing method")
		}
		return []byte(*config.Value), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New(constant.TokenExpiredError)
			}
		}
		return nil, errors.New(constant.IllegalTokenError)
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New(constant.TokenExpiredError)
}

func GetToken(c *gin.Context) string {
	tokenStr := c.Request.Header.Get("Authorization")
	if tokenStr == "" {
		return ""
	}
	return strings.SplitN(tokenStr, " ", 2)[1]
}
