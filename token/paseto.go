package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func NewPasetoMaker(symmetricKey string) (*PasetoMaker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	pm := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return pm, nil
}

func (pm *PasetoMaker) CreateToken(username string, timedur time.Duration) (string, *PlayLoad, error) {
	payload, err := NewPlayLoad(username, timedur)
	if err != nil {
		return "", payload, err
	}

	token, err := pm.paseto.Encrypt(pm.symmetricKey, payload, nil)
	return token, payload, err
}

func (pm *PasetoMaker) ValidToken(token string) (*PlayLoad, error) {
	payload := &PlayLoad{}

	err := pm.paseto.Decrypt(token, pm.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
