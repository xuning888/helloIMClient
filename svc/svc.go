package svc

type CommonSvc struct {
	*ChatSvc
	*UserSvc
}

func NewCommonSvc(chatSvc *ChatSvc, userSvc *UserSvc) *CommonSvc {
	commonSvc := &CommonSvc{
		ChatSvc: chatSvc,
		UserSvc: userSvc,
	}
	return commonSvc
}
