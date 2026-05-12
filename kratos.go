package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func getKratosUserIDFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("ory_kratos_session")
	if err != nil {
		return "", fmt.Errorf("kratos session cookie not found: %w", err)
	}

	client := &http.Client{}
	url = KratosPublicURL + "/sessions/whoami"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("kratos whoami failed: %s", resp.Status)
	}

	var data struct {
		Identity struct {
			ID string `json:"id"`
		} `json:"identity"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Identity.ID, nil
}

func hashRequest(req *VMRequest) (string, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}