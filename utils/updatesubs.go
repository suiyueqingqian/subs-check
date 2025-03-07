package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bestruirui/bestsub/config"
	"github.com/bestruirui/bestsub/utils/log"
)

type versionResponse struct {
	Version string `json:"version"`
}

type providersResponse struct {
	Providers map[string]struct {
		VehicleType string `json:"vehicleType"`
	} `json:"providers"`
}

func makeRequest(method, url string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.GlobalConfig.MihomoApiSecret))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNoContent {
			return nil, nil
		}
		return nil, fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	return body, nil
}

func UpdateSubs() {
	if config.GlobalConfig.MihomoApiUrl == "" {
		log.Warn("MihomoApiUrl not configured, skipping update")
		return
	}

	names, err := getNeedUpdateNames()
	if err != nil {
		log.Error("get need update subs failed: %v", err)
		return
	}

	if err := updateSubs(names); err != nil {
		log.Error("update subs failed: %v", err)
		return
	}
	log.Info("subs updated")
}

func GetVersion() (string, error) {
	url := fmt.Sprintf("%s/version", config.GlobalConfig.MihomoApiUrl)
	body, err := makeRequest(http.MethodGet, url)
	if err != nil {
		return "", err
	}

	var version versionResponse
	if err := json.Unmarshal(body, &version); err != nil {
		return "", fmt.Errorf("parse version info failed: %w", err)
	}
	return version.Version, nil
}

func getNeedUpdateNames() ([]string, error) {
	url := fmt.Sprintf("%s/providers/proxies", config.GlobalConfig.MihomoApiUrl)
	body, err := makeRequest(http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	var response providersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse provider info failed: %w", err)
	}

	var names []string
	for name, provider := range response.Providers {
		if provider.VehicleType == "HTTP" || provider.VehicleType == "File" {
			names = append(names, name)
		}
	}
	return names, nil
}

func updateSubs(names []string) error {
	for _, name := range names {
		url := fmt.Sprintf("%s/providers/proxies/%s", config.GlobalConfig.MihomoApiUrl, name)
		if _, err := makeRequest(http.MethodPut, url); err != nil {
			log.Error("update sub %s failed: %v", name, err)
		}
		log.Info("update sub %s success", name)
	}
	return nil
}
