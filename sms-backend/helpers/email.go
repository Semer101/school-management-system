package helpers

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
)

// SendEmail sends an HTML email via SMTP.
// Port 587 uses STARTTLS, not direct TLS (which is port 465/SMTPS).
// The old code called tls.Dial() which opens a TLS handshake immediately —
// that works on port 465 but is rejected by port 587 (Gmail, SendGrid, etc.).
// The correct sequence for port 587 is:
//  1. Open a plain TCP connection (net.Dial)
//  2. Create an SMTP client
//  3. Upgrade with STARTTLS (client.StartTLS)
//  4. Authenticate and send
func SendEmail(to, subject, body string) error {
	host := os.Getenv("SMTP_HOST") // e.g. smtp.gmail.com
	port := os.Getenv("SMTP_PORT") // e.g. 587
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = user // fall back to SMTP_USER if SMTP_FROM is not set
	}

	if host == "" {
		// No SMTP configured — log and skip silently (dev/test mode)
		fmt.Printf("[EMAIL SKIPPED] To: %s | Subject: %s\n", to, subject)
		return nil
	}

	addr := host + ":" + port

	// Step 1: plain TCP connection on port 587
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	// Step 2: create SMTP client over the plain connection
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	// Step 3: upgrade to TLS with STARTTLS (required on port 587)
	tlsConfig := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("smtp starttls: %w", err)
	}

	// Step 4: authenticate
	auth := smtp.PlainAuth("", user, pass, host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	// Step 5: send the message
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body,
	)
	if _, err = w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp write close: %w", err)
	}

	return client.Quit()
}

// SendGradeNotification emails a student when grades are released.
func SendGradeNotification(studentEmail, studentName, subject string, score float64) {
	body := fmt.Sprintf(`
		<h2>Grade Released</h2>
		<p>Dear %s,</p>
		<p>Your grade for <strong>%s</strong> has been recorded: <strong>%.1f</strong></p>
		<p>Log in to the SMS portal to view your full report card.</p>
	`, studentName, subject, score)
	go SendEmail(studentEmail, "SMS: Grade Released — "+subject, body)
}

// SendAbsenceAlert emails a parent for a single absence (kept for backward compatibility).
func SendAbsenceAlert(parentEmail, studentName, subject, date string) {
	body := fmt.Sprintf(`
		<h2>Absence Alert</h2>
		<p>Dear Parent/Guardian,</p>
		<p><strong>%s</strong> was marked absent in <strong>%s</strong> on %s.</p>
		<p>Please contact the school if you have any questions.</p>
	`, studentName, subject, date)
	go SendEmail(parentEmail, "SMS: Absence Alert for "+studentName, body)
}

// SendAbsenceAlertBulk emails a parent with ALL subjects their child was absent from on a date.
// replaces calling SendAbsenceAlert in a loop (one email per subject) with a single
// aggregated email — prevents spam when a student is absent from multiple subjects.
func SendAbsenceAlertBulk(parentEmail, studentName, subjectList, date string) {
	body := fmt.Sprintf(`
		<h2>Absence Alert</h2>
		<p>Dear Parent/Guardian,</p>
		<p><strong>%s</strong> was marked absent in the following subject(s) on <strong>%s</strong>:</p>
		<p><strong>%s</strong></p>
		<p>Please contact the school if you have any questions.</p>
	`, studentName, date, subjectList)
	go SendEmail(parentEmail, "SMS: Absence Alert for "+studentName+" on "+date, body)
}

// SendBroadcast sends a school-wide announcement email to a list of recipients.
func SendBroadcast(emails []string, title, body string) {
	for _, email := range emails {
		go SendEmail(email, "SMS Announcement: "+title, body)
	}
}
