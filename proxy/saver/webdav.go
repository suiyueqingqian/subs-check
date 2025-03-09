package saver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bestruirui/bestsub/config"
	"github.com/bestruirui/bestsub/utils"
	"github.com/bestruirui/bestsub/utils/log"
)

var (
	webdavMaxRetries = 3
	webdavRetryDelay = 2 * time.Second
)

type WebDAVUploader struct {
	client   *http.Client
	baseURL  string
	username string
	password string
}

func NewWebDAVUploader() *WebDAVUploader {
	return &WebDAVUploader{
		client:   utils.NewHTTPClient(),
		baseURL:  config.GlobalConfig.Save.WebDAVURL,
		username: config.GlobalConfig.Save.WebDAVUsername,
		password: config.GlobalConfig.Save.WebDAVPassword,
	}
}

func UploadToWebDAV(yamlData []byte, filename string) error {
	uploader := NewWebDAVUploader()
	return uploader.Upload(yamlData, filename)
}

func ValiWebDAVConfig() error {
	if config.GlobalConfig.Save.WebDAVURL == "" {
		return fmt.Errorf("webdav URL is not configured")
	}
	if config.GlobalConfig.Save.WebDAVUsername == "" {
		return fmt.Errorf("webdav username is not configured")
	}
	if config.GlobalConfig.Save.WebDAVPassword == "" {
		return fmt.Errorf("webdav password is not configured")
	}
	return nil
}
func (w *WebDAVUploader) Upload(yamlData []byte, filename string) error {
	if err := w.validateInput(yamlData, filename); err != nil {
		return err
	}

	return w.uploadWithRetry(yamlData, filename)
}

func (w *WebDAVUploader) validateInput(yamlData []byte, filename string) error {
	if len(yamlData) == 0 {
		return fmt.Errorf("yaml data is empty")
	}
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	if w.baseURL == "" {
		return fmt.Errorf("webdav URL is not configured")
	}
	return nil
}

func (w *WebDAVUploader) uploadWithRetry(yamlData []byte, filename string) error {
	var lastErr error

	for attempt := 0; attempt < webdavMaxRetries; attempt++ {
		if err := w.doUpload(yamlData, filename); err != nil {
			lastErr = err
			log.Error("webdav upload failed(attempt %d/%d): %v", attempt+1, webdavMaxRetries, err)
			time.Sleep(webdavRetryDelay)
			continue
		}
		log.Info("webdav upload success: %s", filename)
		return nil
	}

	return fmt.Errorf("webdav upload failed, tried %d times: %w", webdavMaxRetries, lastErr)
}

func (w *WebDAVUploader) doUpload(yamlData []byte, filename string) error {
	req, err := w.createRequest(yamlData, filename)
	if err != nil {
		return err
	}
	req.Close = true
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	return w.checkResponse(resp)
}

func (w *WebDAVUploader) createRequest(yamlData []byte, filename string) (*http.Request, error) {
	baseURL := w.baseURL
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	url := baseURL + filename

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(yamlData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.SetBasicAuth(w.username, w.password)
	req.Header.Set("Content-Type", "application/x-yaml")
	return req, nil
}

func (w *WebDAVUploader) checkResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response failed(status code: %d): %w", resp.StatusCode, err)
		}
		return fmt.Errorf("upload failed(status code: %d): %s", resp.StatusCode, string(body))
	}
	return nil
}
