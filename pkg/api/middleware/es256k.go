package middleware

// copied from https://gist.github.com/bnewbold/bc9b97c9b281295da1fa47c03b0b3c69

import (
	"crypto"
	"errors"
	"fmt"

	atcrypto "github.com/bluesky-social/indigo/atproto/crypto"
	"github.com/golang-jwt/jwt/v5"
)

var (
	SigningMethodES256K *SigningMethodAtproto
	SigningMethodES256  *SigningMethodAtproto
)

type SigningMethodAtproto struct {
	alg      string
	hash     crypto.Hash
	toOutSig toOutSig
	sigLen   int
}

type toOutSig func(sig []byte) []byte

func init() {
	SigningMethodES256K = &SigningMethodAtproto{
		alg:      "ES256K",
		hash:     crypto.SHA256,
		toOutSig: toES256K,
		sigLen:   64,
	}
	jwt.RegisterSigningMethod(SigningMethodES256K.Alg(), func() jwt.SigningMethod {
		return SigningMethodES256K
	})
	SigningMethodES256 = &SigningMethodAtproto{
		alg:      "ES256",
		hash:     crypto.SHA256,
		toOutSig: toES256,
		sigLen:   64,
	}
	jwt.RegisterSigningMethod(SigningMethodES256.Alg(), func() jwt.SigningMethod {
		return SigningMethodES256
	})
	fmt.Println("init Completed")
}

// Errors returned on different problems.
var (
	ErrWrongKeyFormat  = errors.New("wrong key type")
	ErrBadSignature    = errors.New("bad signature")
	ErrVerification    = errors.New("signature verification failed")
	ErrFailedSigning   = errors.New("failed generating signature")
	ErrHashUnavailable = errors.New("hasher unavailable")
)

func (sm *SigningMethodAtproto) Verify(signingString string, sig []byte, key any) error {
	pub, ok := key.(atcrypto.PublicKey)
	if !ok {
		return ErrWrongKeyFormat
	}

	if !sm.hash.Available() {
		return ErrHashUnavailable
	}

	if len(sig) != sm.sigLen {
		return ErrBadSignature
	}

	return pub.HashAndVerifyLenient([]byte(signingString), sig)
}

func (sm *SigningMethodAtproto) Sign(signingString string, key any) ([]byte, error) {
	// TODO: implement signatures
	return nil, ErrFailedSigning
}

func (sm *SigningMethodAtproto) Alg() string {
	return sm.alg
}

func toES256K(sig []byte) []byte {
	return sig[:64]
}

func toES256(sig []byte) []byte {
	return sig[:64]
}
