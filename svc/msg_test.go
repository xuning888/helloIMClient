package svc

import (
	"fmt"
	"testing"
)

func TestOrderedMsg_Insert(t *testing.T) {
	orderedMsg := NewOrderedMsg()

	msgs := []*ChatMessage{
		{
			ServerSeq: 2,
		},
		{
			ServerSeq: 5,
		},
		{
			ServerSeq: 3,
		},
		{
			ServerSeq: 4,
		},
		{
			ServerSeq: 1,
		},
	}
	for _, msg := range msgs {
		ordered, minSeq, maxSeq := orderedMsg.Insert(msg)
		fmt.Printf("insert: %v, out: %v, minSeq: %v, maxSeq: %v\n",
			msg, ordered, minSeq, maxSeq)
	}
}
