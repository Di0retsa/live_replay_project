package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"live_replay_project/backend/global"
	"strconv"
	"time"
)

// CustomPayload 自定义载荷，继承原有接口饼附带自己的字段
type CustomPayload struct {
	UserId     uint64
	GrantScope string
	jwt.RegisteredClaims
}

func GenerateToken(uid uint64, subject string, secret string) (string, error) {
	ttl, err := strconv.ParseInt(global.Config.Jwt.TTL, 10, 64)
	if err != nil {
		return "", err
	}
	claims := CustomPayload{
		UserId:     uid,
		GrantScope: subject,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Auth_Server",                                                        // 签发者
			Subject:   subject,                                                              // 签发对象
			Audience:  jwt.ClaimStrings{"PC", "Wechat_Program"},                             // 签发受众
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Second)), // 过期时间
			NotBefore: jwt.NewNumericDate(time.Now().Add(time.Second)),                      // 最早使用时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                       // 签发时间
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return token, err
}

func ParseToken(token string, secret string) (*CustomPayload, error) {
	parseToken, err := jwt.ParseWithClaims(token, &CustomPayload{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parseToken.Claims.(*CustomPayload); ok && parseToken.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
