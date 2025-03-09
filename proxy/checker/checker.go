package checker

import (
	"github.com/bestruirui/bestsub/proxy/info"
)

type Checker struct {
	Proxy *info.Proxy
}

func NewChecker(proxy *info.Proxy) *Checker {
	return &Checker{
		Proxy: proxy,
	}
}

func (c *Checker) Close() {
	if c.Proxy != nil {
		c.Proxy.Close()
	}
}
