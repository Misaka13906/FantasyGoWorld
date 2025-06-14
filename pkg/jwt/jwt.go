package jwt

import (
	"time"

	"github.com/Misaka13906/FantasyGoWorld/internal/config"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

const jwtIssuer = "FantasyGoWorld"

type JWTClaims struct {
	UID string `json:"uid"`
	jwt.StandardClaims
}

func NewClaims(uid string, expirationHours time.Duration) *JWTClaims {
	return &JWTClaims{
		UID: uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expirationHours * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    jwtIssuer,
		},
	}
}

func GenerateToken(claims *JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.Configs.JwtSecretKey))
	if err != nil {
		zap.L().Error("Failed to generate token: ", zap.Error(err))
		return "", err
	}
	return signedToken, nil
}

func ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(config.Configs.JwtSecretKey), nil
	})
	if err != nil {
		zap.L().Error("Failed to parse token: ", zap.Error(err))
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		zap.L().Error("Invalid token claims or token is not valid")
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func IsTokenExpired(tokenString string) (bool, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return false, err
	}
	return claims.ExpiresAt < time.Now().Unix(), nil
}

func RefreshToken(tokenString string, newExpirationHours time.Duration) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	claims.ExpiresAt = time.Now().Add(newExpirationHours * time.Hour).Unix()

	newToken, err := GenerateToken(claims)
	if err != nil {
		return "", err
	}
	return newToken, nil
}
