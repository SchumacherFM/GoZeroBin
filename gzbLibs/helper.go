package gzbLibs

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
)

var (
	// every server restart we'll get another key and every delete link will get invalid
	serverKey = make([]byte, 40)
)

func init() {
	_, err := rand.Read(serverKey)
	if err != nil {
		log.Fatal("crypto rand error:", err)
	}
}

func GetDeleteToken(pasteId string) string {
	mac := hmac.New(sha256.New, serverKey)
	mac.Write([]byte(pasteId))
	pasteMAC := mac.Sum(nil)
	return hex.EncodeToString(pasteMAC)
}

// CheckMAC returns true if messageMAC is a valid HMAC tag for message.
func CheckDeleteToken(message, messageMAC string) bool {
	byteMessageMAC, _ := hex.DecodeString(messageMAC)
	mac := hmac.New(sha256.New, serverKey)
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(byteMessageMAC, expectedMAC)
}
