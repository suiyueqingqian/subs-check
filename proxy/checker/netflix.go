package checker

import (
	"context"
	"net/http"
)

func (c *Checker) NetflixTest() {
	ctx, cancel := context.WithCancel(c.Proxy.Ctx)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.netflix.com/title/81280792", nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	resp, err := c.Proxy.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		c.Proxy.Info.Unlock.Netflix = true
	}
}
