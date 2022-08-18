package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type PlayLoad struct {
	Id           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	IssueAt      time.Time `json:"issueat"`
	TimeExprieAt time.Time `json:"time_exprie_at"`
}

func NewPlayLoad(username string, timedur time.Duration) (*PlayLoad, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return &PlayLoad{}, err
	}

	return &PlayLoad{
		Id:           uid,
		Username:     username,
		IssueAt:      time.Now(),
		TimeExprieAt: time.Now().Add(timedur),
	}, nil
}

func (pl *PlayLoad) Valid() error {

	if ep := time.Now().After(pl.TimeExprieAt); ep {
		return ErrExpiredToken
	}

	return nil
}
