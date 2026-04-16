package email

import (
	"context"
	"fmt"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type MailgunSender struct {
	client   *mailgun.MailgunImpl
	domain   string
	fromAddr string
	fromName string
}

func NewMailgunSender(domain, apiKey, fromAddr, fromName string) *MailgunSender {
	return &MailgunSender{
		client:   mailgun.NewMailgun(domain, apiKey),
		domain:   domain,
		fromAddr: fromAddr,
		fromName: fromName,
	}
}

func (s *MailgunSender) Send(ctx context.Context, input *SendInput) (*SendOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	from := input.From
	if input.FromName != "" {
		from = fmt.Sprintf("%s <%s>", input.FromName, input.From)
	}

	message := s.client.NewMessage(from, input.Subject, input.TextBody, input.To...)

	if input.HTMLBody != "" {
		message.SetHtml(input.HTMLBody)
	}

	for _, cc := range input.CC {
		message.AddCC(cc)
	}

	for _, bcc := range input.BCC {
		message.AddBCC(bcc)
	}

	for _, att := range input.Attachments {
		message.AddBufferAttachment(att.Filename, att.Data)
	}

	for key, value := range input.Headers {
		message.AddHeader(key, value)
	}

	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, id, err := s.client.Send(sendCtx, message)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	return &SendOutput{
		MessageID: id,
		From:      from,
		To:        input.To,
		SentAt:    time.Now(),
		Provider:  ProviderMailgun,
	}, nil
}

func (s *MailgunSender) HealthCheck(ctx context.Context) error {
	return nil
}
