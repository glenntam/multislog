package multislog

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"sync"
)

const queueCapacity = 100

// ErrEmailQueueFull occurs when the email job queue is already full.
var ErrEmailQueueFull = errors.New("email queue full")

// smtpClient contains config settings for a simple SMTP client.
type smtpClient struct {
	Host      string
	Port      string
	Username  string
	Password  string
	Sender    string
	Recipient string

	queue chan emailJob
	wg    sync.WaitGroup
}

// emailJob is a queue of emails.
type emailJob struct {
	subject   string
	body      string
	recipient string
}

// newSMTPClient initializes a simple SMTP client.
func newSMTPClient(host, port, username, password, sender, recipient string) *smtpClient {
	sc := &smtpClient{
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		Sender:    sender,
		Recipient: recipient,
		queue:     make(chan emailJob, queueCapacity),
	}
	sc.wg.Add(1)
	go sc.worker()
	return sc
}

// Close the smtp worker queue gracefully.
func (sc *smtpClient) Close() {
	fmt.Fprintf(os.Stderr, "multislog: attempting to close smtp email queue")
	close(sc.queue)
	sc.wg.Wait()
}

// Send a simple email by sending it to the email queue.
// If the queue is full, the email is dropped.
func (sc *smtpClient) Send(subject, body, recipient string) error {
	select {
	case sc.queue <- emailJob{subject, body, recipient}:
		return nil
	default:
		return ErrEmailQueueFull
	}
}

// sendSync fires smtp.SendMail. It is blocking and meant to be used in a queue.
func (sc *smtpClient) sendSync(subject, body, recipient string) error {
	addr := fmt.Sprintf("%s:%s", sc.Host, sc.Port)
	auth := smtp.PlainAuth("", sc.Username, sc.Password, sc.Host)
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", recipient, subject, body)

	err := smtp.SendMail(addr, auth, sc.Sender, []string{recipient}, []byte(msg))
	if err != nil {
		return ErrEmailQueueFull
	}
	return nil
}

// worker perpetually checks the email queue for emails to send.
func (sc *smtpClient) worker() {
	defer sc.wg.Done()
	for job := range sc.queue {
		_ = sc.sendSync(job.subject, job.body, job.recipient)
	}
}
