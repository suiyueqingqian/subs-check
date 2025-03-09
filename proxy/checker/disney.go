package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Checker) DisneyTest() {
	ctx, cancel := context.WithCancel(c.Proxy.Ctx)
	defer cancel()

	const (
		cookie    = "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Atoken-exchange&latitude=0&longitude=0&platform=browser&subject_token=DISNEYASSERTION&subject_token_type=urn%3Abamtech%3Aparams%3Aoauth%3Atoken-type%3Adevice"
		assertion = `{"deviceFamily":"browser","applicationRuntime":"chrome","deviceProfile":"windows","attributes":{}}`
		authBear  = "Bearer ZGlzbmV5JmJyb3dzZXImMS4wLjA.Cu56AgSfBTDag5NiRA81oLHkDZfu5L3CKadnefEAY84"
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
	)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://disney.api.edge.bamgrid.com/devices", strings.NewReader(assertion))
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", authBear)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Proxy.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var assertionResp map[string]interface{}
	if err := json.Unmarshal(body, &assertionResp); err != nil {
		return
	}

	assertionToken, ok := assertionResp["assertion"].(string)
	if !ok {
		return
	}

	tokenData := strings.Replace(cookie, "DISNEYASSERTION", assertionToken, 1)
	req, err = http.NewRequestWithContext(ctx, "POST", "https://disney.api.edge.bamgrid.com/token", strings.NewReader(tokenData))
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", authBear)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.Proxy.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return
	}

	if errDesc, ok := tokenResp["error_description"].(string); ok && errDesc == "forbidden-location" {
		return
	}

	refreshToken, ok := tokenResp["refresh_token"].(string)
	if !ok {
		return
	}

	gqlQuery := fmt.Sprintf(`{"query":"mutation refreshToken($input: RefreshTokenInput!) {refreshToken(refreshToken: $input) {activeSession {sessionId}}}","variables":{"input":{"refreshToken":"%s"}}}`, refreshToken)

	req, err = http.NewRequestWithContext(ctx, "POST", "https://disney.api.edge.bamgrid.com/graph/v1/device/graphql", strings.NewReader(gqlQuery))
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", authBear)

	resp, err = c.Proxy.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var gqlResp map[string]interface{}
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return
	}

	extensions, ok := gqlResp["extensions"].(map[string]interface{})
	if !ok {
		return
	}

	sdk, ok := extensions["sdk"].(map[string]interface{})
	if !ok {
		return
	}

	session, ok := sdk["session"].(map[string]interface{})
	if !ok {
		return
	}

	inSupportedLocation, _ := session["inSupportedLocation"].(bool)

	c.Proxy.Info.Unlock.Disney = inSupportedLocation
}
