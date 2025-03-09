package checker

import (
	"context"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/bestruirui/bestsub/config"
	"github.com/bestruirui/bestsub/utils/log"
	"github.com/dlclark/regexp2"
)

func (c *Checker) CheckSpeed() {

	if config.GlobalConfig.Check.SpeedSkipName != "" {
		re := regexp2.MustCompile(config.GlobalConfig.Check.SpeedSkipName, regexp2.None)
		match, err := re.MatchString(c.Proxy.Raw["name"].(string))
		if err != nil {
			log.Debug("check speed skip name failed: %v", err)
			return
		}
		if match {
			c.Proxy.Info.SpeedSkip = true
			log.Debug("check speed skip : %v", c.Proxy.Raw["name"])
			return
		}
	}

	speedClient := &http.Client{
		Timeout:   time.Duration(config.GlobalConfig.Check.DownloadTimeout) * time.Second,
		Transport: c.Proxy.Client.Transport,
	}

	successChan := make(chan *http.Response, 1)
	ctx, cancel := context.WithCancel(c.Proxy.Ctx)
	defer cancel()

	var startTime time.Time

	for _, url := range config.GlobalConfig.Check.SpeedTestUrl {
		go func(testUrl string) {
			req, err := http.NewRequestWithContext(ctx, "GET", testUrl, nil)
			if err != nil {
				return
			}

			trace := &httptrace.ClientTrace{
				GotFirstResponseByte: func() {
					startTime = time.Now()
				},
			}
			req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

			resp, err := speedClient.Do(req)
			if err != nil {
				return
			}

			select {
			case successChan <- resp:
			default:
				resp.Body.Close()
			}
		}(url)
	}

	var resp *http.Response
	select {
	case resp = <-successChan:
	case <-time.After(time.Duration(config.GlobalConfig.Check.DownloadTimeout) * time.Second):
		return
	}
	defer resp.Body.Close()

	var totalBytes int64

	limitedReader := &io.LimitedReader{
		R: resp.Body,
		N: int64(config.GlobalConfig.Check.DownloadSize) * 1024 * 1024,
	}

	totalBytes, err := io.Copy(io.Discard, limitedReader)
	if err != nil {
		return
	}

	duration := time.Since(startTime).Milliseconds()
	if duration == 0 {
		duration = 1
	}

	c.Proxy.Info.Speed = int(float64(totalBytes) / 1024 * 1000 / float64(duration))
}
