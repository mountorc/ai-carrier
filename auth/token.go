package auth

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var tokenStore = make(map[string]tokenInfo)

type tokenInfo struct {
	userUUID string
	expires  time.Time
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func verifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(user *User) (string, error) {
	token := uuid.New().String()
	hashedToken := hashToken(token)

	tokenStore[hashedToken] = tokenInfo{
		userUUID: user.UUID,
		expires:  time.Now().Add(24 * time.Hour),
	}

	return token, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

func validateToken(token string) (*User, error) {
	hashedToken := hashToken(token)

	info, exists := tokenStore[hashedToken]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}

	if time.Now().After(info.expires) {
		delete(tokenStore, hashedToken)
		return nil, fmt.Errorf("token expired")
	}

	return getUserByUUID(info.userUUID)
}

func invalidateToken(token string) {
	hashedToken := hashToken(token)
	delete(tokenStore, hashedToken)
}