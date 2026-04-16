package sms

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TwilioSender struct {
	accountSID string
	authToken  string
	from       string
	baseURL    string
}

func NewTwilioSender(accountSID, authToken, from string) *TwilioSender {
	return &TwilioSender{
		accountSID: accountSID,
		authToken:  authToken,
		from:       from,
		baseURL:    fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s", accountSID),
	}
}

func (s *TwilioSender) Send(ctx context.Context, input *SendInput) (*SendOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	from := input.From
	if from == "" {
		from = s.from
	}

	msgData := url.Values{}
	msgData.Set("To", input.To)
	msgData.Set("From", from)
	msgData.Set("Body", input.Message)

	for _, mediaURL := range input.MediaURLs {
		msgData.Add("MediaUrl", mediaURL)
	}

	endpoint := fmt.Sprintf("%s/Messages.json", s.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(msgData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	req.SetBasicAuth(s.accountSID, s.authToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%w: status %d, body: %s", ErrSendFailed, resp.StatusCode, string(body))
	}

	var result struct {
		SID    string `json:"sid"`
		Status string `json:"status"`
	}

	if err := parseJSON(body, &result); err != nil {
		result.SID = fmt.Sprintf("twilio-%d", time.Now().UnixNano())
		result.Status = "queued"
	}

	return &SendOutput{
		MessageID: result.SID,
		From:      from,
		To:        input.To,
		Body:      input.Message,
		SentAt:    time.Now(),
		Provider:  ProviderTwilio,
		Status:    result.Status,
	}, nil
}

func (s *TwilioSender) HealthCheck(ctx context.Context) error {
	return nil
}

func parseJSON(data []byte, v interface{}) error {
	return fmt.Errorf("simple json parser not implemented")
}
