package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/xuning888/helloIMClient/internal/dal/sqllite"
	"github.com/xuning888/helloIMClient/pkg"
)

var (
	ipListPath                   = "/index/iplist"
	allUserPath                  = "/user/allUser"
	allChatPath                  = "/chat/getAllChat"
	lastMessagePath              = "/chat/lastMessage"
	pullOfflineMsgPath           = "/message/pullOfflineMsg"
	getLatestOfflineMessagesPath = "/message/getLatestOfflineMessages"
)

// IpList 服务发现获取长连接公网IP地址
// path: /index/iplist
func IpList(ctx context.Context) ([]string, error) {
	var result pkg.RestResult[[]string]
	var url = baseUrl + ipListPath
	resp, err := restClient.R().SetContext(ctx).SetResult(&result).Get(url)
	if err != nil {
		return nil, fmt.Errorf("iplist 请求失败: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("iplist HTTP错误: %d, 响应: %s", resp.StatusCode(), resp.String())
	}
	// 业务错误检查
	if result.Code != 0 {
		return nil, fmt.Errorf("iplist 业务异常: code=%d, msg=%s", result.Code, result.Msg)
	}
	ips := result.Data
	return ips, nil
}

// Users 拉去所有用户信息
func Users(ctx context.Context) ([]*sqllite.ImUser, error) {
	var result pkg.RestResult[[]*sqllite.ImUser]
	var url = baseUrl + allUserPath
	resp, err := restClient.R().SetContext(ctx).SetResult(&result).Get(url)
	if err != nil {
		return nil, fmt.Errorf("allUser 请求失败: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("allUser HTTP错误: %d, 响应: %s", resp.StatusCode(), resp.String())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("allUser 业务异常: code=%d, msg=%s", result.Code, result.Msg)
	}
	users := result.Data
	return users, nil
}

func GetAllChat(userId int64) ([]*sqllite.ImChat, error) {
	var result pkg.RestResult[[]*sqllite.ImChat]
	var url = baseUrl + allChatPath + fmt.Sprintf("?userId=%d", userId)
	resp, err := restClient.R().SetResult(&result).Get(url)
	if err != nil {
		return nil, fmt.Errorf("GetAllChat 请求失败: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("GetAllChat HTTP错误: %d, 响应: %s", resp.StatusCode(), resp.String())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("GetAllChat 业务异常: code=%d, msg=%s", result.Code, result.Msg)
	}
	chats := result.Data
	return chats, nil
}

func LastMessage(userId, chatId int64, chatType int32) (*sqllite.ChatMessage, error) {
	var result pkg.RestResult[*sqllite.ChatMessage]
	var params = fmt.Sprintf("?userId=%d&chatId=%d&chatType=%d", userId, chatId, chatType)
	var url = baseUrl + lastMessagePath + params
	resp, err := restClient.R().SetResult(&result).Get(url)
	if err != nil {
		return nil, fmt.Errorf("LastMessage 请求失败: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("LastMessage HTTP错误: %d, 响应: %s", resp.StatusCode(), resp.String())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("LastMessage 业务异常: code=%d, msg=%s", result.Code, result.Msg)
	}
	message := result.Data
	return message, nil
}

func PullOfflineMsg(fromUserId int64, chatId int64, chatType int32, minServerSeq, maxServerSeq int64) ([]*sqllite.ChatMessage, error) {
	var result pkg.RestResult[[]*sqllite.ChatMessage]
	params := fmt.Sprintf("?fromUserId=%d&chatId=%d&chatType=%d&minServerSeq=%d&maxServerSeq=%d",
		fromUserId, chatId, chatType, minServerSeq, maxServerSeq)
	url := baseUrl + pullOfflineMsgPath + params
	resp, err := restClient.R().SetResult(&result).Get(url)
	if err != nil {
		return nil, fmt.Errorf("PullOfflineMsg 请求失败: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("PullOfflineMsg HTTP错误: %d, 响应: %s", resp.StatusCode(), resp.String())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("PullOfflineMsg 业务异常: code=%d, msg=%s", result.Code, result.Msg)
	}
	return result.Data, nil
}

func GetLatestOfflineMessages(fromUserId int64, chatId int64, chatType int32, size int32) ([]*sqllite.ChatMessage, error) {
	var result pkg.RestResult[[]*sqllite.ChatMessage]
	params := fmt.Sprintf("?fromUserId=%d&chatId=%d&chatType=%d&size=%d",
		fromUserId, chatId, chatType, size)
	url := baseUrl + getLatestOfflineMessagesPath + params
	resp, err := restClient.R().SetResult(&result).Get(url)
	if err != nil {
		return nil, fmt.Errorf("GetLatestOfflineMessages 请求失败: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("GetLatestOfflineMessages HTTP错误: %d, 响应: %s", resp.StatusCode(), resp.String())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("GetLatestOfflineMessages 业务异常: code=%d, msg=%s", result.Code, result.Msg)
	}
	return result.Data, nil
}
