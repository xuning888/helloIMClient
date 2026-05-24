package im

import "time"

// Options SDK 配置
type Options struct {
	UID              int64         // 用户ID
	Token            string        // 认证token
	ConnectTimeout   time.Duration // 连接超时
	Reconnect        bool          // 是否自动重连
	KeepLiveInterval time.Duration // 心跳间隔
}

func NewOptions() *Options {
	return &Options{
		ConnectTimeout:   time.Second * 5,
		KeepLiveInterval: time.Second * 10,
		Reconnect:        true,
	}
}

type Option func(opt *Options)

func WithUID(uid int64) Option {
	return func(opt *Options) {
		opt.UID = uid
	}
}

func WithToken(token string) Option {
	return func(opt *Options) {
		opt.Token = token
	}
}

func WithConnectTimeout(connectTimeout time.Duration) Option {
	return func(opt *Options) {
		opt.ConnectTimeout = connectTimeout
	}
}

func WithReconnect(reconnect bool) Option {
	return func(opt *Options) {
		opt.Reconnect = reconnect
	}
}

func WithKeepLiveInterval(keepLiveInterval time.Duration) Option {
	return func(opt *Options) {
		opt.KeepLiveInterval = keepLiveInterval
	}
}
