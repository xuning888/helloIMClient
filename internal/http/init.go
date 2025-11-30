package http

import (
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	baseUrl    string
	restClient *resty.Client
)

func Init(serverUrl string, timeout time.Duration) {
	baseUrl = serverUrl
	restClient = resty.New().
		SetBaseURL(baseUrl).
		SetTimeout(timeout).
		SetHeader("Accept", "application/json")
}
