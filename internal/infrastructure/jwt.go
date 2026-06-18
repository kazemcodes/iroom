package infrastructure

import (
	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/pkg/jwt"
)

type JWTProvider struct {
	secret string
}

func NewJWTProvider(secret string) *JWTProvider {
	return &JWTProvider{secret: secret}
}

func (p *JWTProvider) Generate(claims entity.TokenClaims, expiryMinutes int) (string, error) {
	jwtClaims := jwt.Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}
	return jwt.Generate(p.secret, jwtClaims, expiryMinutes)
}

func (p *JWTProvider) Validate(token string) (*entity.TokenClaims, error) {
	jwtClaims, err := jwt.Validate(p.secret, token)
	if err != nil {
		return nil, err
	}
	return &entity.TokenClaims{
		UserID: jwtClaims.UserID,
		Email:  jwtClaims.Email,
		Role:   jwtClaims.Role,
	}, nil
}
