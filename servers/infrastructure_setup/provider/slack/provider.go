// Package slack implements the Provider interface for Slack channel management.
// It uses the Slack Web API with a bot token.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

const slackAPIBase = "https://slack.com/api"

// Config holds credentials for the Slack provider.
type Config struct {
	BotToken string `json:"bot_token"`
}

// Provider implements provider.Provider for Slack.
type Provider struct {
	cfg    Config
	client *http.Client
}

// New creates a Slack provider from decoded credentials.
func New(cfg Config) *Provider {
	return &Provider{cfg: cfg, client: &http.Client{}}
}

func (p *Provider) GetType() string { return "slack" }

func (p *Provider) GetAuthFields() []provider.AuthField {
	return []provider.AuthField{
		{Name: "bot_token", Label: "Bot Token", Type: "password", Required: true,
			Description: "Slack bot token starting with xoxb-. Requires scopes: channels:manage, groups:write, users:read, users:read.email"},
	}
}

func (p *Provider) ValidateCredentials(ctx context.Context) error {
	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := p.callAPI(ctx, "auth.test", map[string]interface{}{}, &resp); err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("slack auth.test failed: %s", resp.Error)
	}
	return nil
}

// CreateResource creates a private Slack channel and invites members.
// It is idempotent: if the channel name is already taken, the existing channel is reused.
func (p *Provider) CreateResource(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	channelName := sanitizeChannelName(input.Name)

	channelID, channelURL, err := p.findOrCreateChannel(ctx, channelName)
	if err != nil {
		return nil, err
	}

	for _, member := range input.Members {
		userID, err := p.lookupUserByEmail(ctx, member.Email)
		if err != nil {
			// Skip users not found in Slack rather than failing the whole operation.
			continue
		}
		if err := p.inviteToChannel(ctx, channelID, userID); err != nil {
			return nil, fmt.Errorf("inviting %s to channel: %w", member.Email, err)
		}
	}

	return &provider.Resource{ExternalID: channelID, ExternalURL: channelURL}, nil
}

// DeleteResource archives a Slack channel by its ID.
func (p *Provider) DeleteResource(ctx context.Context, externalID string) error {
	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := p.callAPI(ctx, "conversations.archive", map[string]interface{}{"channel": externalID}, &resp); err != nil {
		return err
	}
	if !resp.OK && resp.Error != "already_archived" && resp.Error != "channel_not_found" {
		return fmt.Errorf("slack conversations.archive failed: %s", resp.Error)
	}
	return nil
}

// findOrCreateChannel returns the channel ID for the given name, creating it if needed.
func (p *Provider) findOrCreateChannel(ctx context.Context, name string) (string, string, error) {
	var createResp struct {
		OK      bool   `json:"ok"`
		Error   string `json:"error"`
		Channel struct {
			ID string `json:"id"`
		} `json:"channel"`
	}

	if err := p.callAPI(ctx, "conversations.create", map[string]interface{}{
		"name":       name,
		"is_private": true,
	}, &createResp); err != nil {
		return "", "", err
	}

	if createResp.OK {
		channelURL := fmt.Sprintf("https://slack.com/app_redirect?channel=%s", createResp.Channel.ID)
		return createResp.Channel.ID, channelURL, nil
	}

	// Handle idempotency: if name is already taken, look it up.
	if createResp.Error == "name_taken" || createResp.Error == "already_name_taken" {
		return p.lookupChannelByName(ctx, name)
	}

	return "", "", fmt.Errorf("slack conversations.create failed: %s", createResp.Error)
}

// lookupChannelByName finds an existing channel by name using conversations.list.
func (p *Provider) lookupChannelByName(ctx context.Context, name string) (string, string, error) {
	var listResp struct {
		OK       bool   `json:"ok"`
		Error    string `json:"error"`
		Channels []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"channels"`
		ResponseMetadata struct {
			NextCursor string `json:"next_cursor"`
		} `json:"response_metadata"`
	}

	cursor := ""
	for {
		params := map[string]interface{}{
			"types":            "private_channel",
			"exclude_archived": true,
			"limit":            200,
		}
		if cursor != "" {
			params["cursor"] = cursor
		}

		if err := p.callAPI(ctx, "conversations.list", params, &listResp); err != nil {
			return "", "", err
		}
		if !listResp.OK {
			return "", "", fmt.Errorf("slack conversations.list failed: %s", listResp.Error)
		}

		for _, ch := range listResp.Channels {
			if ch.Name == name {
				channelURL := fmt.Sprintf("https://slack.com/app_redirect?channel=%s", ch.ID)
				return ch.ID, channelURL, nil
			}
		}

		if listResp.ResponseMetadata.NextCursor == "" {
			break
		}
		cursor = listResp.ResponseMetadata.NextCursor
	}

	return "", "", fmt.Errorf("slack channel %q not found after name conflict", name)
}

// inviteToChannel invites a user to a channel, ignoring "already in channel" errors.
func (p *Provider) inviteToChannel(ctx context.Context, channelID, userID string) error {
	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := p.callAPI(ctx, "conversations.invite", map[string]interface{}{
		"channel": channelID,
		"users":   userID,
	}, &resp); err != nil {
		return err
	}
	if !resp.OK && resp.Error != "already_in_channel" {
		return fmt.Errorf("slack conversations.invite failed: %s", resp.Error)
	}
	return nil
}

// lookupUserByEmail resolves a Slack user ID from an email address.
func (p *Provider) lookupUserByEmail(ctx context.Context, email string) (string, error) {
	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := p.callAPI(ctx, "users.lookupByEmail", map[string]interface{}{"email": email}, &resp); err != nil {
		return "", err
	}
	if !resp.OK {
		return "", fmt.Errorf("slack users.lookupByEmail failed for %s: %s", email, resp.Error)
	}
	return resp.User.ID, nil
}

// callAPI makes a Slack Web API call using POST with a JSON body.
func (p *Provider) callAPI(ctx context.Context, method string, params map[string]interface{}, result interface{}) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s", slackAPIBase, method), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.cfg.BotToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack %s: HTTP %d: %s", method, resp.StatusCode, body)
	}

	return json.Unmarshal(body, result)
}

// sanitizeChannelName produces a valid Slack channel name: lowercase, max 80 chars, no spaces.
var nonSlackChar = regexp.MustCompile(`[^a-z0-9-_]`)

func sanitizeChannelName(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = nonSlackChar.ReplaceAllString(s, "")
	s = strings.Trim(s, "-")
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}
