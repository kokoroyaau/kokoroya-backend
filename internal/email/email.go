package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"kokoroya-backend/config"
)

const resendAPIURL = "https://api.resend.com/emails"

var ErrSendFailed = errors.New("email: failed to send")

type Service interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}

type service struct {
	apiKey string
	from   string
}

func NewService(cfg *config.Config) Service {
	return &service{
		apiKey: cfg.Email.ResendAPIKey,
		from:   fmt.Sprintf("%s <%s>", cfg.Email.FromName, cfg.Email.FromAddress),
	}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

type sendErrorResponse struct {
	Message string `json:"message"`
}

func (s *service) Send(ctx context.Context, to, subject, htmlBody string) error {
	body, err := json.Marshal(sendRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return parseSendError(respBody)
	}

	return nil
}

func parseSendError(respBody []byte) error {
	var errResp sendErrorResponse
	if json.Unmarshal(respBody, &errResp) == nil && errResp.Message != "" {
		return fmt.Errorf("%w: %s", ErrSendFailed, errResp.Message)
	}
	return ErrSendFailed
}
