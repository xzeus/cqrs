package ioc

import (
	j "github.com/vizidrix/jose"
)

// Certificate represents a public certificate for the app
type Certificate struct {
	KeyName string
	Data    []byte // PEM-encoded X.509 certificate
}

type Crypto interface {
	// TODO: Implement these as it makes sense
	// Map to ae features for key management, can map to JWK?
	// PublicCertificates retrieves the public certificates for the app
	// They can be used to verify a signature returned by SignByte
	//PublicCertificates() ([]Certificate, error)
	// SignBytes signs bytes using a private key unique to your application
	//SignBytes(bytes []byte) (keyName string, signature []byte, err error)
	DecodeToken([]byte, ...j.TokenModifier) (*j.TokenDef, error)
	EncodeToken(...j.TokenModifier) ([]byte, error)
	Hash32(key []byte) int32
	Hash64(key []byte) int64
	CrcKeyHash(key []byte) int64
	RandInt32() int32
	RandInt64() int64
}
