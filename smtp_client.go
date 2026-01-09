package multislog

import (
	"fmt"
	"net/smtp"
)

// smtpClient contains config settings for a simple SMTP client.
type smtpClient struct {
	Host      string
	Port      string
	Username  string
	Password  string
	Sender    string
	Recipient string
}

// newSMTPClient initializes a simple SMTP client.
func newSMTPClient(host, port, username, password, sender, recipient string) *smtpClient {
	return &smtpClient{
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		Sender:    sender,
		Recipient: recipient,
	}
}

// Send a simple email based on previously set config settings.
func (sc *smtpClient) Send(subject, body, recipient string) error {
	addr := fmt.Sprintf("%s:%s", sc.Host, sc.Port)
	auth := smtp.PlainAuth("", sc.Username, sc.Password, sc.Host)
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", recipient, subject, body)
	err := smtp.SendMail(addr, auth, sc.Sender, []string{recipient}, []byte(msg))
	if err != nil {
		return fmt.Errorf("SMTP Client couldn't send mail. Error: %w", err)
	}
	return nil
}
