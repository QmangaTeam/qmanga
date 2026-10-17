package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Config struct {
	Address, Host, Username, Password, From string
	RequireTLS                              bool
}
type Mailer struct{ config Config }

func New(config Config) *Mailer { return &Mailer{config: config} }
func (mailer *Mailer) Send(ctx context.Context, to, subject, body string) error {
	if strings.ContainsAny(to+subject+mailer.config.From, "\r\n") {
		return fmt.Errorf("invalid email header")
	}
	dialer := net.Dialer{Timeout: 5 * time.Second}
	connection, failure := dialer.DialContext(ctx, "tcp", mailer.config.Address)
	if failure != nil {
		return failure
	}
	defer connection.Close()
	deadline := time.Now().Add(10 * time.Second)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	if failure = connection.SetDeadline(deadline); failure != nil {
		return failure
	}
	client, failure := smtp.NewClient(connection, mailer.config.Host)
	if failure != nil {
		return failure
	}
	defer client.Close()
	if supported, _ := client.Extension("STARTTLS"); supported {
		if failure = client.StartTLS(&tls.Config{ServerName: mailer.config.Host, MinVersion: tls.VersionTLS12}); failure != nil {
			return failure
		}
	} else if mailer.config.RequireTLS {
		return fmt.Errorf("SMTP server requires TLS support")
	}
	if mailer.config.Username != "" {
		if failure = client.Auth(smtp.PlainAuth("", mailer.config.Username, mailer.config.Password, mailer.config.Host)); failure != nil {
			return failure
		}
	}
	if failure = client.Mail(mailer.config.From); failure != nil {
		return failure
	}
	if failure = client.Rcpt(to); failure != nil {
		return failure
	}
	writer, failure := client.Data()
	if failure != nil {
		return failure
	}
	message := "From: " + mailer.config.From + "\r\nTo: " + to + "\r\nSubject: " + mime.QEncoding.Encode("UTF-8", subject) + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + strings.ReplaceAll(body, "\n", "\r\n")
	if _, failure = writer.Write([]byte(message)); failure != nil {
		_ = writer.Close()
		return failure
	}
	if failure = writer.Close(); failure != nil {
		return failure
	}
	return client.Quit()
}
