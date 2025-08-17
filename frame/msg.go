package frame

import (
	pb "github.com/xuning888/helloIMClient/proto"
	"google.golang.org/protobuf/proto"
)

func MakeEchoRequest(msg string, seq int32) (*Frame, error) {
	echo := &pb.EchoRequest{
		Msg: msg,
	}
	marshal, err := proto.Marshal(echo)
	if err != nil {
		return nil, err
	}
	return &Frame{
		Header: NewMsgHeader(1, seq, int(pb.CmdId_CMD_ID_ECHO), len(marshal)),
		Body:   marshal,
	}, nil
}

func MakeC2cSendMessage(from, to, content string, contentType int, seq int32) (*Frame, error) {
	req := &pb.C2CSendRequest{
		From:        from,
		To:          to,
		Content:     content,
		ContentType: int32(contentType),
	}
	marshal, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	return &Frame{
		Header: NewMsgHeader(1, seq, int(pb.CmdId_CMD_ID_C2CSEND), len(marshal)),
		Body:   marshal,
	}, nil
}
