package transport

import "time"

type Option func(opt *Options)

type Options struct {
	serverUrl   string
	httpTimeout time.Duration
	maxAttempts int           // 发送消息时的最大重试次数
	lingerMs    time.Duration // 发送消息的间隔时间
	initSeq     int32         // 初始seq
}

// WithServerUrl 设置IM服务端WebAPI地址
func WithServerUrl(serverUrl string) Option {
	return func(opt *Options) {
		opt.serverUrl = serverUrl
	}
}

// WithHttpTimeout http的全局超时时间
func WithHttpTimeout(timeout time.Duration) Option {
	return func(opt *Options) {
		opt.httpTimeout = timeout
	}
}

// WithMaxAttempts 发送消息的最大重试次数
func WithMaxAttempts(maxAttempts int) Option {
	return func(opt *Options) {
		opt.maxAttempts = maxAttempts
	}
}

// WithLingerMs 没有消息时最大的等待时间
func WithLingerMs(lingerMs time.Duration) Option {
	return func(opt *Options) {
		opt.lingerMs = lingerMs
	}
}

func WithInitSeq(seq int32) Option {
	return func(opt *Options) {
		opt.initSeq = seq
	}
}

func loadOptions(options ...Option) *Options {
	opts := &Options{}
	for _, option := range options {
		option(opts)
	}
	return opts
}
