package utamaru

import (
	"fmt"
	"time"
	"rand"
	"encoding/base64"
	"crypto/hmac"
)

func GetUniqId(remoteAddr, userAgent string) string {
	now := time.Seconds()
	r := rand.New(rand.NewSource(now))
	msg := fmt.Sprintf("%s%s%d%d", remoteAddr, userAgent, now, r.Int())
	h := hmac.NewSHA1([]byte("utamaru"))
	h.Write([]byte(msg))
	return base64.StdEncoding.EncodeToString(h.Sum())
}
