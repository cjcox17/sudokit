package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	useTLS   bool
	fromAddr string
	fromName string
}

func NewSMTPSender(host string, port int, username, password, fromAddr, fromName string, useTLS bool) *SMTPSender {
	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		useTLS:   useTLS,
		fromAddr: fromAddr,
		fromName: fromName,
	}
}

func (s *SMTPSender) Send(ctx context.Context, input *SendInput) (*SendOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	from := input.From
	if input.FromName != "" {
		from = fmt.Sprintf("%s <%s>", input.FromName, input.From)
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(input.To, ", ")))
	if len(input.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(input.CC, ", ")))
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", input.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	if len(input.Attachments) > 0 {
		boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))

		if input.HTMLBody != "" {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
			buf.WriteString(input.HTMLBody)
			buf.WriteString("\r\n")
		} else {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
			buf.WriteString(input.TextBody)
			buf.WriteString("\r\n")
		}

		for _, att := range input.Attachments {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", att.ContentType))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Filename))
			b64 := base64.StdEncoding.EncodeToString(att.Data)
			buf.WriteString(b64)
			buf.WriteString("\r\n")
		}
		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		if input.HTMLBody != "" {
			buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
			buf.WriteString(input.HTMLBody)
		} else {
			buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
			buf.WriteString(input.TextBody)
		}
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	recipients := append(input.To, input.CC...)
	recipients = append(recipients, input.BCC...)

	var err error
	if s.useTLS {
		err = s.sendTLS(addr, auth, input.From, recipients, buf.Bytes())
	} else {
		err = smtp.SendMail(addr, auth, input.From, recipients, buf.Bytes())
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	return &SendOutput{
		MessageID: fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), s.host),
		From:      input.From,
		To:        input.To,
		SentAt:    time.Now(),
		Provider:  ProviderSMTP,
	}, nil
}

func (s *SMTPSender) sendTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.host,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	return w.Close()
}

func (s *SMTPSender) HealthCheck(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.host,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderNotConfig, err)
	}
	conn.Close()
	return nil
}
