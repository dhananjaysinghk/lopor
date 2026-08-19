package email

import (
	"fmt"
	"log"
	"net/smtp"
)

type Mailer struct {
	SMTPHost string
	SMTPPort string
	From     string
}

func NewMailer(host, port, from string) *Mailer {
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "1025" // Mailpit default
	}
	if from == "" {
		from = "noreply@lopor.ai"
	}
	return &Mailer{
		SMTPHost: host,
		SMTPPort: port,
		From:     from,
	}
}

// SendInvitationEmail dispatches an HTML workspace invitation email via SMTP / Mailpit
func (m *Mailer) SendInvitationEmail(recipientEmail, orgName, inviteLink string) error {
	subject := fmt.Sprintf("Subject: You're invited to join %s on Lopor AI Workspace\n", orgName)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<div style="font-family: sans-serif; padding: 20px; background: #0f1117; color: #f3f4f6; border-radius: 12px;">
			<h2 style="color: #6366f1;">Join %s on Lopor AI</h2>
			<p>You have been invited to collaborate in <strong>%s</strong> on Lopor AI Workspace.</p>
			<p><a href="%s" style="display: inline-block; padding: 10px 20px; background: #4f46e5; color: white; border-radius: 8px; text-decoration: none; font-weight: bold;">Accept Invitation</a></p>
			<p style="font-size: 12px; color: #9ca3af;">Or copy link: %s</p>
		</div>
	`, orgName, orgName, inviteLink, inviteLink)

	msg := []byte(subject + mime + body)
	addr := fmt.Sprintf("%s:%s", m.SMTPHost, m.SMTPPort)

	// Send via local SMTP / Mailpit
	err := smtp.SendMail(addr, nil, m.From, []string{recipientEmail}, msg)
	if err != nil {
		log.Printf("[Email Mailer Log]: Pre-rendered email for %s (Mailpit fallback): %v", recipientEmail, err)
		return nil // Non-blocking in dev mode
	}

	log.Printf("[Email Mailer] Sent invitation email to %s successfully!", recipientEmail)
	return nil
}
