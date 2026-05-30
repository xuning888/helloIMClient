package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/xuning888/helloIMClient/im/dbhelper"
	"github.com/xuning888/helloIMClient/pkg"
)

type WebApiClient struct {
	cli *resty.Client
}

func NewRestClient(baseUrl string, timeout time.Duration) *WebApiClient {
	cli := resty.New().
		SetBaseURL(baseUrl).
		SetTimeout(timeout).
		SetHeader("Accept", "application/json")
	return &WebApiClient{cli: cli}
}

func (r *WebApiClient) GetAddr(ctx context.Context) ([]string, error) {
	return r.IpList(ctx)
}

func (r *WebApiClient) IpList(ctx context.Context) ([]string, error) {
	var result pkg.RestResult[[]string]
	resp, err := r.cli.R().SetContext(ctx).SetResult(&result).Get(ipListPath)
	if err != nil {
		return nil, fmt.Errorf("iplist 请求失败: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("iplist HTTP错误: %d, 响应: %s", resp.StatusCode(), resp.String())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("iplist 业务异常: code=%d, msg=%s", result.Code, result.Msg)
	}
	return result.Data, nil
}

func (r *WebApiClient) Users(ctx context.Context) ([]*dbhelper.ImUser, error) {
	var result pkg.RestResult[[]*dbhelper.ImUser]
	resp, err := r.cli.R().SetContext(ctx).SetResult(&result).Get(allUserPath)
	if err != nil {
		return nil, fmt.Errorf("allUser 请求失败: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("allUser HTTP错误: %d, 响应: %s", resp.StatusCode(), resp.String())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("allUser 业务异常: code=%d, msg=%s", result.Code, result.Msg)
	}
	return result.Data, nil
}

func (r *WebApiClient) GetAllChat(ctx context.Context, userId int64) ([]*dbhelper.ImChat, error) {
	var result pkg.RestResult[[]*dbhelper.ImChat]
	resp, err := r.cli.R().SetContext(ctx).SetResult(&result).Get(allChatPath + fmt.Sprintf("?userId=%d", userId))
	if err != nil {
		return nil, fmt.Errorf("GetAllChat 请求失败: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("GetAllChat HTTP错误: %d, 响应: %s", resp.StatusCode(), resp.String())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("GetAllChat 业务异常: code=%d, msg=%s", result.Code, result.Msg)
	}
	return result.Data, nil
}

func (r *WebApiClient) LastMessage(ctx context.Context, userId, chatId int64, chatType int32) (*dbhelper.ChatMessage, error) {
	var result pkg.RestResult[*dbhelper.ChatMessage]
	params := fmt.Sprintf("?userId=%d&chatId=%d&chatType=%d", userId, chatId, chatType)
	resp, err := r.cli.R().SetContext(ctx).SetResult(&result).Get(lastMessagePath + params)
	if err != nil {
		return nil, fmt.Errorf("LastMessage 请求失败: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("LastMessage HTTP错误: %d, 响应: %s", resp.StatusCode(), resp.String())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("LastMessage 业务异常: code=%d, msg=%s", result.Code, result.Msg)
	}
	return result.Data, nil
}

func (r *WebApiClient) PullOfflineMsg(
	ctx context.Context, fromUserId int64, chatId int64, chatType int32,
	minServerSeq, maxServerSeq int64) ([]*dbhelper.ChatMessage, error) {
	var result pkg.RestResult[[]*dbhelper.ChatMessage]
	params := fmt.Sprintf("?fromUserId=%d&chatId=%d&chatType=%d&minServerSeq=%d&maxServerSeq=%d",
		fromUserId, chatId, chatType, minServerSeq, maxServerSeq)
	resp, err := r.cli.R().SetContext(ctx).SetResult(&result).Get(pullOfflineMsgPath + params)
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

func (r *WebApiClient) GetLatestOfflineMessages(ctx context.Context,
	fromUserId int64, chatId int64, chatType int32, size int32) ([]*dbhelper.ChatMessage, error) {
	var result pkg.RestResult[[]*dbhelper.ChatMessage]
	params := fmt.Sprintf("?fromUserId=%d&chatId=%d&chatType=%d&size=%d",
		fromUserId, chatId, chatType, size)
	resp, err := r.cli.R().SetContext(ctx).SetResult(&result).Get(getLatestOfflineMessagesPath + params)
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
