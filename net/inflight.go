package net

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/xuning888/helloIMClient/frame"
)

type OnRetry func(f *frame.Frame) error

type WhenComplete func(request, response *frame.Frame, err error)

type inflightItem struct {
	frame         *frame.Frame   // 绑定的request
	whenComplete  WhenComplete   // 回调函数
	expireAt      time.Time      // 过期时间
	timer         *time.Timer    // 定时器
	queue         *inflightQueue // 飞行队列
	retryCount    int            // 重试次数
	firstSendTime time.Time      // 首次发送时间
	interval      time.Duration  // 重试的间隔时间
	nextRetryTime time.Time      // 下一次发送时间
	lock          sync.Mutex     // lock
}

type inflightQueue struct {
	messages     sync.Map
	totalTimeout time.Duration
	maxRetries   int
	interval     time.Duration
	onRetry      OnRetry
}

func makeInflightQueue(totalTimeout time.Duration, maxRetries int, interval time.Duration, onRetry OnRetry) *inflightQueue {
	return &inflightQueue{
		totalTimeout: totalTimeout,
		interval:     interval,
		onRetry:      onRetry,
		maxRetries:   maxRetries,
	}
}

func (iq *inflightQueue) Put(request *frame.Frame, onComplete WhenComplete) bool {
	key := request.Key()
	now := time.Now()

	item := &inflightItem{
		frame:         request,
		expireAt:      now.Add(iq.totalTimeout),
		whenComplete:  onComplete,
		queue:         iq,
		firstSendTime: now,
		interval:      iq.interval,
		nextRetryTime: now.Add(iq.interval),
		lock:          sync.Mutex{},
	}

	_, exists := iq.messages.LoadOrStore(key, item)
	if exists {
		return false
	}

	item.timer = time.AfterFunc(iq.interval, item.checkRetryOrTimeout)
	return true
}

func (iq *inflightQueue) Ack(reply *frame.Frame) {
	key := reply.Key()
	//log.Printf("消息ACK, key: %v\n", key)
	if requestItem, exists := iq.messages.LoadAndDelete(key); exists {
		if item, ok := requestItem.(*inflightItem); ok {
			request := item.frame
			item.safeStop()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("panic in whenComplete, error: %v\n", r)
				}
			}()
			item.whenComplete(request, reply, nil) // 回调处理器
		}
	}
}

func (iq *inflightQueue) remove(request *frame.Frame) {
	key := request.Key()
	iq.messages.Delete(key)
}

func (iq *inflightQueue) RemoveAndStop(request *frame.Frame) {
	key := request.Key()
	if requestItem, exists := iq.messages.LoadAndDelete(key); exists {
		if item, ok := requestItem.(*inflightItem); ok {
			item.safeStop()
		}
	}
}

func (item *inflightItem) checkRetryOrTimeout() {
	now := time.Now()

	if item.retryCount >= item.queue.maxRetries {
		item.handleTimeout()
		return
	}

	// 当前时间大于过期时间
	if now.After(item.expireAt) {
		item.handleTimeout()
		return
	}

	if now.After(item.nextRetryTime) {
		if err := item.retry(); err != nil {
			item.safeStop()
		}
	}
}

func (item *inflightItem) retry() (err error) {
	item.lock.Lock()
	item.retryCount++

	newInterval := item.interval * 2
	if item.firstSendTime.Add(newInterval).After(item.expireAt) {
		newInterval = item.expireAt.Sub(time.Now())
	}
	item.interval = newInterval
	item.nextRetryTime = time.Now().Add(item.interval)

	item.lock.Unlock()

	err = item.callback()
	item.timer.Reset(item.nextRetryTime.Sub(time.Now()))

	// 如果发现重试次数已经没了直接失败
	if item.retryCount >= item.queue.maxRetries {
		item.stopAndCallback()
	}
	return
}

func (item *inflightItem) callback() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in callback, error: %v\n", r)
		}
	}()
	request := item.frame
	log.Printf("飞行队列重发消息, key: %v, retryCount: %d\n", request.Key(), item.retryCount)
	if err := item.queue.onRetry(request); err != nil {
		return err
	}
	return nil
}

func (item *inflightItem) handleTimeout() {
	item.lock.Lock()
	defer item.lock.Unlock()

	item.stopAndCallback()
}

func (item *inflightItem) stopAndCallback() {
	item.stop()
	cmdId, seq := item.frame.Header.CmdId, item.frame.Header.Seq
	err := fmt.Errorf("message timed out [cmdId=%d, seq=%d], retryCount: %v", cmdId, seq, item.retryCount)
	item.whenComplete(item.frame, nil, err)
}

func (item *inflightItem) safeStop() {
	item.lock.Lock()
	defer item.lock.Unlock()
	item.stop()
}

func (item *inflightItem) stop() {
	if item.timer != nil {
		item.timer.Stop()
	}
	item.queue.remove(item.frame)
}
