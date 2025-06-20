// routine/email.go
package routine

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"backend/config"
	generaldb "backend/db/gen/general"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	queries *generaldb.Queries
	dialer  *gomail.Dialer
}

type AlertEmail struct {
	Host     string
	Severity string
	Content  string
	Time     string
}

func NewEmailService(queries *generaldb.Queries) *EmailService {
	// Create SMTP dialer
	dialer := gomail.NewDialer(
		config.AppConfig.SMTPHost,
		config.AppConfig.SMTPPort,
		config.AppConfig.SMTPUsername,
		config.AppConfig.SMTPPassword,
	)

	return &EmailService{
		queries: queries,
		dialer:  dialer,
	}
}

func (es *EmailService) SendAlertEmail(host, severity, content string) error {
	// Check if SMTP is configured
	if config.AppConfig.SMTPUsername == "" || config.AppConfig.SMTPPassword == "" {
		log.Printf("📧 SMTP not configured, skipping email for alert: %s", content)
		return nil
	}

	// Get admin users from database
	adminUsers, err := es.getAdminUsers()
	if err != nil {
		log.Printf("❌ Failed to get admin users: %v", err)
		return err
	}

	if len(adminUsers) == 0 {
		log.Printf("⚠️ No admin users found to send alert email")
		return nil
	}

	// Create email content
	alertEmail := AlertEmail{
		Host:     host,
		Severity: severity,
		Content:  content,
		Time:     time.Now().Format("2006-01-02 15:04:05"),
	}

	// Send email to all admin users
	return es.sendToAdmins(adminUsers, alertEmail)
}

// ✅ Fixed: Email is string type, not sql.NullString
func (es *EmailService) getAdminUsers() ([]generaldb.ListUsersRow, error) {
	// Get all users with admin role
	users, err := es.queries.ListUsers(context.Background())
	if err != nil {
		return nil, err
	}

	var adminUsers []generaldb.ListUsersRow
	for _, user := range users {
		// ✅ Check if user has admin role and valid email (string type)
		if user.Role == "admin" && user.Email != "" {
			adminUsers = append(adminUsers, user)
		}
	}

	return adminUsers, nil
}

// ✅ Updated function signature to use ListUsersRow
func (es *EmailService) sendToAdmins(adminUsers []generaldb.ListUsersRow, alert AlertEmail) error {
	// Prepare recipient list
	var recipients []string
	for _, user := range adminUsers {
		// ✅ Email is string, so use directly
		if user.Email != "" {
			recipients = append(recipients, user.Email)
		}
	}

	if len(recipients) == 0 {
		log.Printf("⚠️ No admin email addresses found")
		return nil
	}

	// Create email message
	message := gomail.NewMessage()

	// Set headers
	message.SetHeader("From", config.AppConfig.SMTPFrom)
	message.SetHeader("To", recipients...)
	message.SetHeader("Subject", fmt.Sprintf("[SNSMS Alert] %s - %s",
		strings.ToUpper(alert.Severity), alert.Host))

	// Create HTML email body
	htmlBody := es.createAlertEmailHTML(alert)
	message.SetBody("text/html", htmlBody)

	// Create plain text alternative
	textBody := es.createAlertEmailText(alert)
	message.AddAlternative("text/plain", textBody)

	// Send email
	if err := es.dialer.DialAndSend(message); err != nil {
		log.Printf("❌ Failed to send alert email: %v", err)
		return err
	}

	log.Printf("📧 Alert email sent successfully to %d admin(s): %s",
		len(recipients), strings.Join(recipients, ", "))
	return nil
}

func (es *EmailService) createAlertEmailHTML(alert AlertEmail) string {
	severityColor := es.getSeverityColor(alert.Severity)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>SNSMS Alert</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #0f172a; color: white; padding: 20px; border-radius: 8px 8px 0 0; }
        .content { background: #f8f9fa; padding: 20px; border: 1px solid #dee2e6; }
        .footer { background: #6c757d; color: white; padding: 15px; border-radius: 0 0 8px 8px; text-align: center; }
        .alert-box { background: %s; color: white; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .detail-row { margin: 10px 0; }
        .label { font-weight: bold; color: #495057; }
        .value { color: #212529; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚨 SNSMS Alert Notification</h1>
        </div>
        <div class="content">
            <div class="alert-box">
                <h2>%s ALERT</h2>
                <p><strong>Host:</strong> %s</p>
            </div>
            
            <div class="detail-row">
                <span class="label">Alert Details:</span><br>
                <span class="value">%s</span>
            </div>
            
            <div class="detail-row">
                <span class="label">Time:</span>
                <span class="value">%s</span>
            </div>
            
            <div class="detail-row">
                <span class="label">Severity:</span>
                <span class="value" style="color: %s; font-weight: bold;">%s</span>
            </div>
            
            <hr>
            <p><em>This is an automated alert from your SNSMS (Server Network Management Suite) system.</em></p>
            <p><em>Please log in to your dashboard to view more details and manage this alert.</em></p>
        </div>
        <div class="footer">
            <p>SNSMS - Server Network Management Suite</p>
            <p>Generated at %s</p>
        </div>
    </div>
</body>
</html>`,
		severityColor,
		strings.ToUpper(alert.Severity),
		alert.Host,
		alert.Content,
		alert.Time,
		severityColor,
		strings.ToUpper(alert.Severity),
		alert.Time)
}

func (es *EmailService) createAlertEmailText(alert AlertEmail) string {
	return fmt.Sprintf(`
SNSMS Alert Notification
========================

ALERT: %s
Host: %s
Time: %s
Severity: %s

Details:
%s

---
This is an automated alert from your SNSMS (Server Network Management Suite) system.
Please log in to your dashboard to view more details and manage this alert.

SNSMS - Server Network Management Suite
Generated at %s
`,
		strings.ToUpper(alert.Severity),
		alert.Host,
		alert.Time,
		strings.ToUpper(alert.Severity),
		alert.Content,
		alert.Time)
}

func (es *EmailService) getSeverityColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "#dc3545" // Red
	case "warning":
		return "#fd7e14" // Orange
	case "info":
		return "#0dcaf0" // Light blue
	default:
		return "#6c757d" // Gray
	}
}
