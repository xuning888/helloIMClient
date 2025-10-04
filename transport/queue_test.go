package transport

import "testing"

func TestGet(t *testing.T) {
	queue := newSyncQueue(10, 1)
	item := queue.Get()
	if item == nil {
		return
	}
}
