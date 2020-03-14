package registry

import "time"

// option patten to register plugin

type Options struct {
	// address of the register center
	Addrs [] string

	// timeout of communicate with register center
	Timeout time.Duration

	// the Register path
	// example:  /xxx_example/app/kuaishou/service_a/190.12.32.1:8001
	RegisterPath string

	// heart beat
	HeartBeat int64
}


type Option func(opt *Options)

// init the timeout duration
func WithTimeout(timeout time.Duration) Option {
	return func(opt *Options) {
		opt.Timeout = timeout
	}
}

// init the addr
func WithAddrs(addrs []string) Option {
	return func(opt *Options) {
		opt.Addrs = addrs
	}
}

// init the path
func WithRegisterPath(path string) Option {
	return func(opt *Options) {
		opt.RegisterPath = path
	}
}

// init the heart beat
func WithHeartBeat(heartBeat int64) Option {
	return func(opt *Options) {
		opt.HeartBeat = heartBeat
	}
}
