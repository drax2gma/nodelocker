package x

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost for bcrypt hashing (between 10 and 14 recommended for production)
	BcryptCost = 12

	// Hash version prefixes
	bcryptPrefix = "$2a$"
	sha1Prefix   = "sha1$"

	// Legacy SHA1 salts
	preSalt  = "68947b1f416c3a5655e1ff9e7c7935f6"
	postSalt = "5f09dd9c81596ea3cc93ce0df58e26d8"
)

// HashPassword creates a bcrypt hash of the password with version prefix
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a password against a hashed password, handling both old and new formats
func CheckPassword(password, hashedPassword string) bool {
	// Check if it's a bcrypt hash
	if strings.HasPrefix(hashedPassword, bcryptPrefix) {
		err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		return err == nil
	}

	// Fall back to SHA1 for legacy passwords
	h := sha1.New()
	h.Write([]byte(preSalt + password + postSalt))
	sha1Hash := fmt.Sprintf("%x", h.Sum(nil))
	return sha1Hash == hashedPassword
}

// NeedsUpgrade checks if the password hash needs to be upgraded
func NeedsUpgrade(hashedPassword string) bool {
	return !strings.HasPrefix(hashedPassword, bcryptPrefix)
}
