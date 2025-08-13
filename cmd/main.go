package main

import (
	"log"
	"time"

	"github.com/xuning888/helloIMClient/internal/im"
	"github.com/xuning888/helloIMClient/internal/net"
)

var Me = &im.User{}

func main() {
	client := net.NewImClient("127.0.0.1:9300")
	err := client.Start()
	if err != nil {
		log.Fatal(err)
	}
	msg := []byte("hello world\n")
	f := &net.Frame{
		Header: &net.MsgHeader{
			HeaderLength:  int32(net.DefaultHeaderSize),
			ClientVersion: 1,
			Seq:           0,
			CmdId:         1,
			BodyLength:    int32(len(msg)),
		},
		Body: msg,
	}
	client.Process()
	go func() {
		for {
			err = client.SendFrame(f)
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second)
		}
	}()
	select {}
}
