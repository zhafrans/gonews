package auth

import (
	"fmt"
	"gonews/config"
	"gonews/internal/core/domain/entity"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Jwt interface {
	GenerateToken(data *entity.JwtData) (string, int64, error)
	VerifyAccessToken(token string) (*entity.JwtData, error)
}

type Options struct {
	SignInKey string
	Issuer    string
}

func (o *Options) GenerateToken(data *entity.JwtData) (string, int64, error) {
	now := time.Now().Local()
	expiredAt := now.Add(24 * time.Hour)

	data.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiredAt),
		Issuer:    o.Issuer,
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	signed, err := token.SignedString([]byte(o.SignInKey))
	if err != nil {
		return "", 0, err
	}

	return signed, expiredAt.Unix(), nil
}

func (o *Options) VerifyAccessToken(token string) (*entity.JwtData, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &entity.JwtData{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(o.SignInKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(*entity.JwtData); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func NewJwt(cfg *config.Config) Jwt {
	return &Options{
		SignInKey: cfg.App.JwtSecretKey,
		Issuer:    cfg.App.JwtIssuer,
	}
}
