package utamaru

import (
	"io"
	"fmt"
	"math/big"
	"time"
	"crypto/rand"
	"encoding/base64"
	"crypto/sha1"
)

func GetUniqId(remoteAddr, userAgent string) string {
	now := time.Now().Unix()
	n, _ := rand.Int(rand.Reader, big.NewInt(999999))
	msg := fmt.Sprintf("%s%s%d%d", remoteAddr, userAgent, now, int(n.Int64()))
	h := sha1.New()
	io.WriteString(h, msg)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
