package svc

type CommonSvc struct {
	*ChatSvc
	*UserSvc
	*MsgSvc
}

func NewCommonSvc(users []*User, chats []*Chat) *CommonSvc {
	commonSvc := &CommonSvc{
		ChatSvc: newChatSvc(chats),
		UserSvc: newUserSvc(users),
	}
	return commonSvc
}
