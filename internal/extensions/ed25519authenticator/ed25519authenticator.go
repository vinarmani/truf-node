package ed25519authenticator

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

type Ed25519Authenticator struct{}

// Verify verifies the signature of a message
func (Ed25519Authenticator) Verify(sender, msg, signature []byte) error {
	if !ed25519.Verify(sender, msg, signature) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// Identifier returns the hex encoding of the public key
func (Ed25519Authenticator) Identifier(sender []byte) (string, error) {
	return hex.EncodeToString(sender), nil
}
