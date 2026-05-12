package helpers

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to, subject, body string) error {
	host := os.Getenv("SMTP_HOST") // smtp.gmail.com
	port := os.Getenv("SMTP_PORT") // 587
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if host == "" {
		fmt.Printf("[EMAIL SKIPPED] To: %s\n", to)
		return nil
	}

	auth := smtp.PlainAuth("", user, pass, host)
	tlsConfig := &tls.Config{ServerName: host}

	conn, err := tls.Dial("tcp", host+":"+port, tlsConfig)
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	if err = client.Auth(auth); err != nil {
		return err
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", from, to, subject, body)
	w.Write([]byte(msg))
	w.Close()
	return client.Quit()
}

// SendGradeNotification emails a student when grades are released (FE-18)
func SendGradeNotification(studentEmail string, studentName string, subject string, score float64) {
	emailBody := fmt.Sprintf(`
		<h2>Grade Released</h2>
		<p>Dear %s,</p>
		<p>Your grade for <strong>%s</strong> has been recorded: <strong>%.1f</strong></p>
		<p>Log in to the SMS portal to view your full report card.</p>
	`, studentName, subject, score)

	go SendEmail(studentEmail, "SMS: Grade Released — "+subject, emailBody)
}

// SendAbsenceAlert emails a parent when their child is absent (FE-08)
func SendAbsenceAlert(parentEmail string, studentName string, subject string, date string) {
	emailBody := fmt.Sprintf(`
		<h2>Absence Alert</h2>
		<p>Dear Parent/Guardian,</p>
		<p><strong>%s</strong> was marked absent in <strong>%s</strong> on %s.</p>
		<p>Please contact the school if you have any questions.</p>
	`, studentName, subject, date)

	go SendEmail(parentEmail, "SMS: Absence Alert for "+studentName, emailBody)
}

// SendBroadcast sends a school-wide announcement (FE-19)
func SendBroadcast(emails []string, title string, body string) {
	for _, email := range emails {
		go SendEmail(email, "SMS Announcement: "+title, body)
	}
}
