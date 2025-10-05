package transport

import "testing"

func TestGet(t *testing.T) {
	queue := newSyncQueue(10, 1)
	item := queue.get()
	if item == nil {
		return
	}
}
