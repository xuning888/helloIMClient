package main

import (
	"log"

	"github.com/xuning888/helloIMClient/im"
	"github.com/xuning888/helloIMClient/net"
)

var Me = &im.User{}

func main() {
	cli := net.NewImClient("127.0.0.1:9300")
	if err := cli.Start(); err != nil {
		log.Fatal(err)
	}
}
