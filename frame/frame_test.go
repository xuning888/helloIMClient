package frame

import "testing"

func BenchmarkFrame_ToBytes(b *testing.B) {
	body := []byte("hello world")
	h := NewMsgHeader(1, 1, 1, len(body))
	f := &Frame{
		Header: h,
		Body:   body,
	}
	for i := 0; i < b.N; i++ {
		f.ToBytes()
	}
}

func BenchmarkFrame_ToBytesV2(b *testing.B) {
	body := []byte("hello world")
	h := NewMsgHeader(1, 1, 1, len(body))
	f := &Frame{
		Header: h,
		Body:   body,
	}
	for i := 0; i < b.N; i++ {
		f.ToBytesV2()
	}
}
