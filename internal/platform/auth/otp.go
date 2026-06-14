package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// OTPDigits is the length of a one-time passcode.
const OTPDigits = 6

// NewOTP returns a cryptographically-random 6-digit numeric passcode (zero-padded,
// e.g. "004271"). The caller hashes it (bcrypt) for storage and emails the
// plaintext; it is verified by comparison, never looked up by value.
func NewOTP() (string, error) {
	max := big.NewInt(1_000_000) // 000000–999999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", OTPDigits, n.Int64()), nil
}
