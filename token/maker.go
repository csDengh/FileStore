package token

import (
	"time"
)

type Maker interface {
	CreateToken(username string, time time.Duration) (string, *PlayLoad, error)
	ValidToken(token string) (*PlayLoad, error)
}
