package payload

import (
	im "github.com/xuning888/helloIMClient/internal/proto"
)

// NewTextMessage 构造文本消息
func NewTextMessage(content string, at bool, atUid []string) *im.Payload {
	payload := &im.Payload{
		PayloadType: im.PayloadType_TEXT,
		At:          at,
		AtUid:       atUid,
		Content: &im.Payload_Text{
			Text: NewTextPayload(content),
		},
	}
	return payload
}

// NewImageMessage 构造图片消息
func NewImageMessage(imgUrl string, at bool, atUid []string) *im.Payload {
	payload := &im.Payload{
		PayloadType: im.PayloadType_IMAGE,
		At:          at,
		AtUid:       atUid,
		Content: &im.Payload_Image{
			Image: NewImagePayload(imgUrl),
		},
	}
	return payload
}

// NewFileMessage 构造文件消息
func NewFileMessage(filename, fileUrl string, at bool, atUid []string) *im.Payload {
	payload := &im.Payload{
		PayloadType: im.PayloadType_FILE,
		At:          at,
		AtUid:       atUid,
		Content: &im.Payload_File{
			File: NewFilePayload(filename, fileUrl),
		},
	}
	return payload
}

// NewReceiptMessage 构造已读回执
func NewReceiptMessage(msgId, serverSeq int64) *im.Payload {
	payload := &im.Payload{
		PayloadType: im.PayloadType_RECEIPT,
		At:          false,
		AtUid:       make([]string, 0),
		Content: &im.Payload_Receipt{
			Receipt: NewReceiptPayload(msgId, serverSeq),
		},
	}
	return payload
}

func NewTextPayload(content string) *im.TextPayload {
	return &im.TextPayload{
		Content: content,
	}
}

func NewImagePayload(imageUrl string) *im.ImagePayload {
	return &im.ImagePayload{
		ImageUrl: imageUrl,
	}
}

func NewFilePayload(filename, fileUrl string) *im.FilePayload {
	return &im.FilePayload{
		Filename: filename,
		FileUrl:  fileUrl,
	}
}

// ExtractContent 从 Payload 中提取内容和类型
func ExtractContent(p *im.Payload) (string, int32) {
	switch p.GetPayloadType() {
	case im.PayloadType_TEXT:
		if t := p.GetText(); t != nil {
			return t.GetContent(), int32(p.GetPayloadType())
		}
	case im.PayloadType_IMAGE:
		if img := p.GetImage(); img != nil {
			return img.GetImageUrl(), int32(p.GetPayloadType())
		}
	case im.PayloadType_FILE:
		if f := p.GetFile(); f != nil {
			return f.GetFileUrl(), int32(p.GetPayloadType())
		}
	}
	return "", int32(p.GetPayloadType())
}

func NewReceiptPayload(msgId, serverSeq int64) *im.ReceiptPayload {
	return &im.ReceiptPayload{
		Receipts: []*im.ReceiptPayload_Data{
			{
				MsgId:     msgId,
				ServerSeq: serverSeq,
			},
		},
	}
}
