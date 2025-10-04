package http

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/xuning888/helloIMClient/pkg"
	"net/http"
	"time"
)

var (
	ipListEndpoint = "/index/iplist"
)

type Client struct {
	baseUrl string
	client  *resty.Client
}

func NewClient(baseUrl string, timeout time.Duration) *Client {
	return &Client{
		baseUrl: baseUrl,
		client: resty.New().
			SetBaseURL(baseUrl).
			SetTimeout(timeout).
			SetHeader("Accept", "application/json"),
	}
}

// IpList 服务发现获取长连接公网IP地址
// path: /index/iplist
func (hc *Client) IpList(ctx context.Context) ([]string, error) {
	var result pkg.RestResult[[]string]
	var url = hc.baseUrl + ipListEndpoint
	resp, err := hc.client.R().SetContext(ctx).SetResult(&result).Get(url)
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
