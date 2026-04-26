package oss

import (
	"crypto/rand"
	"fmt"
)

type TokenInfo struct {
	Token        string `json:"token"`
	UuidAutoAuth string `json:"uuid_autoAuth"`
	BasePath     string `json:"basePath"`
	CreatedAt    int64  `json:"createdAt"`
	ExpiresAt    int64  `json:"expiresAt"`
}

type TokenConfig struct {
	Tokens []TokenInfo `json:"tokens"`
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] & 0x3F) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func generateShortToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}