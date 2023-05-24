package v1

import (
	"fmt"
	"sync"
	"time"
)

// 互斥锁，确保并发安全
var mutex sync.Mutex

// 订单计数器
var orderCount int64 = 0

var lastTime int64 = 0

// 生成订单ID
func generateOrderID() string {
	now := time.Now()
	prefix := now.Format("20060102150405")
	mutex.Lock()
	if lastTime != now.Unix() {
		orderCount = 0
		lastTime = now.Unix()
	}
	orderCount++
	count := orderCount
	mutex.Unlock()
	return fmt.Sprintf("%s%03d", prefix, count)
}
