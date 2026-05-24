package payload

import (
	"github.com/xuning888/helloIMClient/im/proto"
)

// NewTextMessage 构造文本消息
func NewTextMessage(content string, at bool, atUid []string) *helloim_proto.Payload {
	payload := &helloim_proto.Payload{
		PayloadType: helloim_proto.PayloadType_TEXT,
		At:          at,
		AtUid:       atUid,
		Content: &helloim_proto.Payload_Text{
			Text: NewTextPayload(content),
		},
	}
	return payload
}

// NewImageMessage 构造图片消息
func NewImageMessage(imgUrl string, at bool, atUid []string) *helloim_proto.Payload {
	payload := &helloim_proto.Payload{
		PayloadType: helloim_proto.PayloadType_IMAGE,
		At:          at,
		AtUid:       atUid,
		Content: &helloim_proto.Payload_Image{
			Image: NewImagePayload(imgUrl),
		},
	}
	return payload
}

// NewFileMessage 构造文件消息
func NewFileMessage(filename, fileUrl string, at bool, atUid []string) *helloim_proto.Payload {
	payload := &helloim_proto.Payload{
		PayloadType: helloim_proto.PayloadType_FILE,
		At:          at,
		AtUid:       atUid,
		Content: &helloim_proto.Payload_File{
			File: NewFilePayload(filename, fileUrl),
		},
	}
	return payload
}

// NewReceiptMessage 构造已读回执
func NewReceiptMessage(msgId, serverSeq int64) *helloim_proto.Payload {
	payload := &helloim_proto.Payload{
		PayloadType: helloim_proto.PayloadType_RECEIPT,
		At:          false,
		AtUid:       make([]string, 0),
		Content: &helloim_proto.Payload_Receipt{
			Receipt: NewReceiptPayload(msgId, serverSeq),
		},
	}
	return payload
}

func NewTextPayload(content string) *helloim_proto.TextPayload {
	return &helloim_proto.TextPayload{
		Content: content,
	}
}

func NewImagePayload(imageUrl string) *helloim_proto.ImagePayload {
	return &helloim_proto.ImagePayload{
		ImageUrl: imageUrl,
	}
}

func NewFilePayload(filename, fileUrl string) *helloim_proto.FilePayload {
	return &helloim_proto.FilePayload{
		Filename: filename,
		FileUrl:  fileUrl,
	}
}

// ExtractContent 从 Payload 中提取内容和类型
func ExtractContent(p *helloim_proto.Payload) (string, int32) {
	switch p.GetPayloadType() {
	case helloim_proto.PayloadType_TEXT:
		if t := p.GetText(); t != nil {
			return t.GetContent(), int32(p.GetPayloadType())
		}
	case helloim_proto.PayloadType_IMAGE:
		if img := p.GetImage(); img != nil {
			return img.GetImageUrl(), int32(p.GetPayloadType())
		}
	case helloim_proto.PayloadType_FILE:
		if f := p.GetFile(); f != nil {
			return f.GetFileUrl(), int32(p.GetPayloadType())
		}
	}
	return "", int32(p.GetPayloadType())
}

func NewReceiptPayload(msgId, serverSeq int64) *helloim_proto.ReceiptPayload {
	return &helloim_proto.ReceiptPayload{
		Receipts: []*helloim_proto.ReceiptPayload_Data{
			{
				MsgId:     msgId,
				ServerSeq: serverSeq,
			},
		},
	}
}
