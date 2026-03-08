package coreRequests

import (
	"io"
	"net/http"
	"time"

	"github.com/prompt-edu/prompt/servers/team_allocation/utils"
	log "github.com/sirupsen/logrus"
)

// Code Duplication because the other function shall move in the library soon
func sendRequest(method, subURL, authHeader string, body io.Reader) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	coreURL := utils.GetCoreUrl()
	requestURL := coreURL + subURL
	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		log.WithFields(log.Fields{
			"method": method,
			"url":    requestURL,
		}).Error("Error creating request:", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"method": method,
			"url":    requestURL,
		}).Error("Error creating request:", err)
		return nil, err
	}

	return resp, nil
}
