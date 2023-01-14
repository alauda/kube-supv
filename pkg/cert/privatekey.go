package cert

import (
	"crypto/rand"
	"crypto/rsa"
)

const (
	DefaultRSAKeyBits   = 2048
	DefaultRSACAKeyBits = 2048
)

func GenerateRSAKey(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}
