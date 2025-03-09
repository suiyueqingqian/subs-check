package checker

import (
	"context"
	"net/http"
	"time"
)

func (c *Checker) AliveTest(url string, expectedStatus int) {
	ctx, cancel := context.WithCancel(c.Proxy.Ctx)
	defer cancel()

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return
	}
	resp, err := c.Proxy.Client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == expectedStatus {
		c.Proxy.Info.Alive = true
		c.Proxy.Info.Delay = uint16(time.Since(start) / time.Millisecond)
	}
}
