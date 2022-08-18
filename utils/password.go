package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pwd string) (string, error) {
	hashedpwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash pwd faild %s", err)
	}
	return string(hashedpwd), nil
}

func ConfirmPwd(hashpwd, pwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashpwd), []byte(pwd))
}
