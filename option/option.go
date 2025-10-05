package option

import "time"

type Option func(opt *Options)

type Options struct {
	ServerUrl   string
	HttpTimeout time.Duration
	MaxAttempts int           // 发送消息时的最大重试次数
	LingerMs    time.Duration // 发送消息的间隔时间
	InitSeq     int32         // 初始seq
}

// WithServerUrl 设置IM服务端WebAPI地址
func WithServerUrl(serverUrl string) Option {
	return func(opt *Options) {
		opt.ServerUrl = serverUrl
	}
}

// WithHttpTimeout http的全局超时时间
func WithHttpTimeout(timeout time.Duration) Option {
	return func(opt *Options) {
		opt.HttpTimeout = timeout
	}
}

// WithMaxAttempts 发送消息的最大重试次数
func WithMaxAttempts(maxAttempts int) Option {
	return func(opt *Options) {
		opt.MaxAttempts = maxAttempts
	}
}

// WithLingerMs 没有消息时最大的等待时间
func WithLingerMs(lingerMs time.Duration) Option {
	return func(opt *Options) {
		opt.LingerMs = lingerMs
	}
}

func WithInitSeq(seq int32) Option {
	return func(opt *Options) {
		opt.InitSeq = seq
	}
}

func LoadOptions(options ...Option) *Options {
	opts := &Options{}
	for _, option := range options {
		option(opts)
	}
	return opts
}
