package core

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateRoomID 生成6位数字的房间ID
func GenerateRoomID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
