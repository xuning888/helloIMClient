package net

import (
	"log"
	"testing"
	"time"

	pb "github.com/xuning888/helloIMClient/internal/proto"
	"google.golang.org/protobuf/proto"
)

func Test_echo(t *testing.T) {
	client := NewImClient("127.0.0.1:9300")
	err := client.Start()
	if err != nil {
		log.Fatal(err)
	}
	req := &pb.EchoRequest{
		Msg: "hello world",
	}
	bytes, err := proto.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}
	f := &Frame{
		Header: &MsgHeader{
			HeaderLength:  int32(DefaultHeaderSize),
			ClientVersion: 1,
			Seq:           0,
			CmdId:         1,
			BodyLength:    int32(len(bytes)),
		},
		Body: bytes,
	}
	err = client.SendFrame(f)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second)
	go func() {
		for reply := range client.replies {
			if reply.Header.CmdId == int32(pb.CmdId_CMD_ID_ECHO) {
				response := &pb.EchoResponse{}
				_ = proto.Unmarshal(reply.Body, response)
				log.Printf("receive: %s\n", response)
			}
		}
	}()
	client.Stop()
}
