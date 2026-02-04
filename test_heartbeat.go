package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("测试心跳机制...")
	
	ticker := time.NewTicker(30 * time.Second)
	stop := make(chan bool)
	
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("💓 发送心跳")
			case <-stop:
				ticker.Stop()
				log.Printf("🛑 心跳停止")
				return
			}
		}
	}()
	
	// 运行2分钟测试
	time.Sleep(120 * time.Second)
	stop <- true
	
	fmt.Println("✅ 心跳测试完成")
}
