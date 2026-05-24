package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xuning888/helloIMClient/im"
	"github.com/xuning888/helloIMClient/im/payload"
	"github.com/xuning888/helloIMClient/im/protocol/send"
	"github.com/xuning888/helloIMClient/pkg/logger"
)

var (
	serverUrl    string
	targetUser   int64
	startUserId  int64
	numUsers     int
	totalPerUser int
)

func init() {
	flag.StringVar(&serverUrl, "serverUrl", "http://127.0.0.1:8087", "服务器地址")
	flag.Int64Var(&targetUser, "target", 0, "目标用户ID")
	flag.Int64Var(&startUserId, "from", 100, "起始用户ID")
	flag.IntVar(&numUsers, "users", 10, "模拟用户数")
	flag.IntVar(&totalPerUser, "n", 1000, "每用户消息数")
}

func main() {
	flag.Parse()
	if targetUser == 0 {
		log.Fatal("请输入目标用户 target")
	}

	if err := logger.InitLogger(); err != nil {
		log.Fatal(err)
	}

	// 统计
	var success atomic.Int64
	var fail atomic.Int64
	var totalLatency atomic.Int64

	start := time.Now()
	var wg sync.WaitGroup

	for u := 0; u < numUsers; u++ {
		uid := startUserId + int64(u)
		sdk, err := im.New(serverUrl,
			im.WithUID(uid),
			im.WithConnectTimeout(time.Second*10),
			im.WithReconnect(false),
		)
		if err != nil {
			log.Printf("user %d: create sdk failed: %v", uid, err)
			fail.Add(int64(totalPerUser))
			continue
		}

		ctx := context.Background()
		if err := sdk.Connect(ctx); err != nil {
			log.Printf("user %d: connect failed: %v", uid, err)
			fail.Add(int64(totalPerUser))
			continue
		}

		wg.Add(1)
		go func(sdk *im.Client, uid int64) {
			defer wg.Done()
			defer sdk.Disconnect(context.Background())
			for i := 0; i < totalPerUser; i++ {
				p := payload.NewTextMessage(fmt.Sprintf("msg %d from uid %d", i, uid), false, nil)
				msg := send.NewSendMsg(uid, targetUser, 1, p, 0, 0)
				reqStart := time.Now()
				_, err := sdk.SendMessage(context.Background(), msg)
				latency := time.Since(reqStart).Microseconds()
				if err != nil {
					fail.Add(1)
				} else {
					success.Add(1)
					totalLatency.Add(latency)
				}
			}
		}(sdk, uid)
	}
	wg.Wait()

	elapsed := time.Since(start)

	succ := success.Load()
	f := fail.Load()
	lat := totalLatency.Load()
	total := numUsers * totalPerUser

	fmt.Println()
	fmt.Println("========== Multi-User Benchmark ==========")
	fmt.Printf("Users:             %d\n", numUsers)
	fmt.Printf("Target:            %d\n", targetUser)
	fmt.Printf("Per-user msgs:     %d\n", totalPerUser)
	fmt.Printf("Total messages:    %d\n", total)
	fmt.Printf("Success:           %d\n", succ)
	fmt.Printf("Failed:            %d\n", f)
	fmt.Printf("Duration:          %v\n", elapsed.Round(time.Millisecond))
	if succ > 0 {
		fmt.Printf("Throughput:        %.2f msg/s\n", float64(succ)/elapsed.Seconds())
		fmt.Printf("Avg latency:       %.2f ms\n", float64(lat)/float64(succ)/1000)
	}
	fmt.Println("============================================")
}
