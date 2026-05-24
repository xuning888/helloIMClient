package transport

import (
	"fmt"

	"github.com/xuning888/helloIMClient/internal/proto"
	"github.com/xuning888/helloIMClient/protocol"
	"google.golang.org/protobuf/proto"
)

type AuthRequest struct {
	*helloim_proto.AuthRequest
}

func (r *AuthRequest) CmdId() int32 {
	return int32(helloim_proto.CmdId_CMD_ID_AUTH)
}

type AuthResponse struct {
	*helloim_proto.AuthResponse
	f *protocol.Frame
}

func (r *AuthResponse) CmdId() int32  { return int32(helloim_proto.CmdId_CMD_ID_AUTH) }
func (r *AuthResponse) MsgSeq() int32 { return r.f.Header.Seq }

func NewAuthRequest(userId int64, userType int32, token string) *AuthRequest {
	return &AuthRequest{
		AuthRequest: &helloim_proto.AuthRequest{
			Uid:      fmt.Sprintf("%d", userId),
			UserType: userType,
			Token:    token,
		},
	}
}

func decodeAuth(frame *protocol.Frame) (protocol.Message, error) {
	resp := &helloim_proto.AuthResponse{}
	if err := proto.Unmarshal(frame.Body, resp); err != nil {
		return nil, err
	}
	return &AuthResponse{AuthResponse: resp, f: frame}, nil
}

func init() {
	protocol.RegisterDecoder(int32(helloim_proto.CmdId_CMD_ID_AUTH), decodeAuth)
}
