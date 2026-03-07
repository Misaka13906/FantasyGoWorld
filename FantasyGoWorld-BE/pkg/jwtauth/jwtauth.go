package jwtauth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint `json:"uid"`
	jwt.RegisteredClaims
}

var (
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenInvalid     = errors.New("token is invalid")
	ErrTokenNotYetValid = errors.New("token not yet valid")
	ErrTokenMalformed   = errors.New("token is malformed")
)

func GenerateToken(uid uint, secret string, expire int64) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Duration(expire) * time.Second)

	claims := Claims{
		UserID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			NotBefore: jwt.NewNumericDate(nowTime),
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString([]byte(secret))

	return token, err
}

func ParseToken(token string, secret string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		} else if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotYetValid
		}
		return nil, ErrTokenInvalid
	}

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, ErrTokenInvalid
}
