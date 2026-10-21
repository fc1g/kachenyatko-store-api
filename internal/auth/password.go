package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

// dummyHash is compared against when the user does not exist,
// so that login takes the same time whether the email is registered or not.
var dummyHash = mustHash("dummy-password-for-timing")

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return hash, nil
}

func ComparePassword(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("compare password and hash: %w", err)
	}

	return match, nil
}

func CompareDummyPassword(password string) {
	_, _ = ComparePassword(password, dummyHash)
}

func mustHash(password string) string {
	hash, err := HashPassword(password)
	if err != nil {
		panic(err)
	}
	return hash
}
