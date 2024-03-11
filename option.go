package csync

import (
	"errors"
	"net/url"
	"os"
)

type config struct {
	Component      Component
	DumpPath       string
	User           string
	Password       string
	TargetUrl      string
	IntervalSecond int
	GetTagFunc     GetTagFunc
	NotifyCh       chan<- NotifyData
}

func newDefaultConfig() *config {
	return &config{
		DumpPath:       "/tmp",
		IntervalSecond: 60,
		GetTagFunc: func() string {
			hostname, _ := os.Hostname()
			return hostname
		},
	}
}

func (c *config) validate() error {
	if c.TargetUrl == "" {
		return errors.New("target url is required")
	}
	if c.User == "" || c.Password == "" {
		return errors.New("basic auth is required")
	}
	if c.Component == "" {
		return errors.New("component is required")
	}
	if c.NotifyCh == nil {
		return errors.New("notify channel is required")
	}
	return nil
}

type optionFunc func(*config) error

func (f optionFunc) Apply(c *config) error {
	return f(c)
}

type Option interface {
	Apply(*config) error
}

func WithComponentName(name Component) Option {
	return optionFunc(func(c *config) error {
		c.Component = name
		return nil
	})
}

func WithDumpPath(path string) Option {
	return optionFunc(func(c *config) error {
		c.DumpPath = path
		return nil
	})
}

func WithBasicAuth(user, password string) Option {
	return optionFunc(func(c *config) error {
		c.User = user
		c.Password = password
		return nil
	})
}

func WithIntervalSecond(interval int) Option {
	return optionFunc(func(c *config) error {
		if interval <= 0 {
			return errors.New("interval second must be greater than 0")
		}
		c.IntervalSecond = interval
		return nil
	})
}

func WithTagFunc(f GetTagFunc) Option {
	return optionFunc(func(c *config) error {
		c.GetTagFunc = f
		return nil
	})
}

func WithTargetUrl(u string) Option {
	return optionFunc(func(c *config) error {
		_, err := url.ParseRequestURI(u)
		if err != nil {
			return errors.New("invalid url")
		}
		c.TargetUrl = u
		return nil
	})
}

func WithNotifyCh(ch chan<- NotifyData) Option {
	return optionFunc(func(c *config) error {
		if ch == nil {
			return errors.New("notify channel is required")
		}
		c.NotifyCh = ch
		return nil
	})
}
