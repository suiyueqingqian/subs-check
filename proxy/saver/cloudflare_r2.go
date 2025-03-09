package saver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bestruirui/bestsub/config"
	"github.com/bestruirui/bestsub/utils"
	"github.com/bestruirui/bestsub/utils/log"
)

const (
	maxRetries    = 3
	retryInterval = 2 * time.Second
)

type KVPayload struct {
	Filename string `json:"filename"`
	Value    string `json:"value"`
}

type R2Uploader struct {
	client    *http.Client
	workerURL string
	token     string
}

func NewR2Uploader() *R2Uploader {
	return &R2Uploader{
		client:    utils.NewHTTPClient(),
		workerURL: config.GlobalConfig.Save.WorkerURL,
		token:     config.GlobalConfig.Save.WorkerToken,
	}
}

func UploadToR2Storage(yamlData []byte, filename string) error {
	uploader := NewR2Uploader()
	return uploader.Upload(yamlData, filename)
}

func ValiR2Config() error {
	if config.GlobalConfig.Save.WorkerURL == "" {
		return fmt.Errorf("worker url is not configured")
	}
	if config.GlobalConfig.Save.WorkerToken == "" {
		return fmt.Errorf("worker token is not configured")
	}
	return nil
}

func (r *R2Uploader) Upload(yamlData []byte, filename string) error {
	if err := r.validateInput(yamlData, filename); err != nil {
		return err
	}

	payload := KVPayload{
		Filename: filename,
		Value:    string(yamlData),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSON encoding failed: %w", err)
	}

	return r.uploadWithRetry(jsonData, filename)
}

func (r *R2Uploader) validateInput(yamlData []byte, filename string) error {
	if len(yamlData) == 0 {
		return fmt.Errorf("yaml data is empty")
	}
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	if r.workerURL == "" || r.token == "" {
		return fmt.Errorf("worker configuration is incomplete")
	}
	return nil
}

func (r *R2Uploader) uploadWithRetry(jsonData []byte, filename string) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := r.doUpload(jsonData); err != nil {
			lastErr = err
			log.Error("upload failed(attempt %d/%d): %v", attempt+1, maxRetries, err)
			time.Sleep(retryInterval)
			continue
		}
		log.Info("upload success: %s", filename)
		return nil
	}

	return fmt.Errorf("upload failed, tried %d times: %w", maxRetries, lastErr)
}

func (r *R2Uploader) doUpload(jsonData []byte) error {
	req, err := r.createRequest(jsonData)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	return r.checkResponse(resp)
}

func (r *R2Uploader) createRequest(jsonData []byte) (*http.Request, error) {
	url := fmt.Sprintf("%s/storage?token=%s", r.workerURL, r.token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (r *R2Uploader) checkResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response failed(status code: %d): %w", resp.StatusCode, err)
		}
		return fmt.Errorf("upload failed(status code: %d): %s", resp.StatusCode, string(body))
	}
	return nil
}
