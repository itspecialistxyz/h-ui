package service

import (
	"errors"
	"h-ui/dao"
	"h-ui/model/bo"
	"h-ui/model/constant"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

const TokenExpireDuration = time.Hour * 24

type MyClaims struct {
	AccountBo bo.AccountBo `json:"account"`
	jwt.RegisteredClaims
}

func GenToken(accountBo bo.AccountBo) (string, error) {
	c := MyClaims{
		AccountBo: accountBo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpireDuration)),
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
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New(constant.TokenExpiredError)
		}
		// For other JWT validation errors (e.g. malformed, invalid signature, not valid yet)
		if errors.Is(err, jwt.ErrTokenMalformed) ||
			errors.Is(err, jwt.ErrTokenSignatureInvalid) ||
			errors.Is(err, jwt.ErrTokenNotValidYet) ||
			errors.Is(err, jwt.ErrInvalidKey) || // Example of another specific error
			errors.Is(err, jwt.ErrInvalidKeyType) { // Example of another specific error
			return nil, errors.New(constant.IllegalTokenError)
		}
		// For any other errors not specifically handled, return a generic illegal token error
		logrus.Errorf("Unhandled JWT parsing error: %v", err)
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
