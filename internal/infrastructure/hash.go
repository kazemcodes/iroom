package infrastructure

import "github.com/iroom/iroom/internal/pkg/hash"

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	return hash.Hash(password)
}

func (h *BcryptHasher) Check(password, hashStr string) bool {
	return hash.Check(password, hashStr)
}
