package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	MixSecretKeyLen = 32
)

type JwtMaker struct {
	SecrertKey string
}

func NewJwtMaker(secretKey string) (*JwtMaker, error) {
	if len(secretKey) < MixSecretKeyLen {
		return nil, fmt.Errorf("secretkey not lower than %d", MixSecretKeyLen)
	}
	return &JwtMaker{SecrertKey: secretKey}, nil
}

func (jm *JwtMaker) CreateToken(username string, timedur time.Duration) (string, *PlayLoad, error) {
	body, err := NewPlayLoad(username, timedur)
	if err != nil {
		return "", body, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, body)

	token, err := jwtToken.SignedString([]byte(jm.SecrertKey))
	return token, body, err
}

func (jm *JwtMaker) ValidToken(token string) (*PlayLoad, error) {
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return []byte(jm.SecrertKey), ErrInvalidToken
		}

		return []byte(jm.SecrertKey), nil
	}

	ptoken, err := jwt.ParseWithClaims(token, &PlayLoad{}, keyfunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	pl, ok := ptoken.Claims.(*PlayLoad)
	if !ok {
		return nil, ErrInvalidToken
	}
	return pl, nil
}
