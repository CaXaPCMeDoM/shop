package hash

import (
	"golang.org/x/crypto/bcrypt"
)

type SHA1 struct {
	salt string
}

func NewSHA1(salt string) *SHA1 {
	return &SHA1{salt: salt}
}

func (h *SHA1) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password+h.salt), bcrypt.DefaultCost)
	return string(bytes), err
}
