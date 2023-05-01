package mails

import (
	"fmt"
	"net/smtp"
	"os"

	a "github.com/maximotejeda/helpers/mails/auth"
)

var (
	senderEmail = os.Getenv("AUTHEMAILUSERNAME")
	pwd         = os.Getenv("AUTHEMAILPWD")
	smtpServer  = os.Getenv("SMTPSERVER")
	smtpPort    = os.Getenv("SMTPPORT")
)

// SendEmail: decide the email kind format and send SendEmail
// too many resposabilities
func SendEmail(mail, token, kind, code, host string) {
	auth := smtp.PlainAuth("", senderEmail, pwd, "smtp.gmail.com")
	to := []string{mail}
	msg := []byte{}
	toBody := fmt.Sprintf("To: %s\r\n", mail)
	subject := ""
	body := ""
	switch kind {
	// need to take this out of here
	case "recover":
		subject = "Subject: Recover your account \r\n"
		body = a.RecoverEmailBody(token, mail, code, host)
		fmt.Println(body)
	case "activate":
		subject = "Subject: Activate your account \r\n"
		body = a.ActivateEmailBody(token, host)
	}

	msg = formatEmail(toBody, subject, body)
	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, senderEmail, to, msg)
	if err != nil {
		panic(err)
	}
}

// formatMailHTML return mimetype to enable html interpreter on emails
func formatMailHTML() string {
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	return mime
}

// formatEmail: set email headers body subject and to
func formatEmail(to, subject, msg string) []byte {
	mimeType := formatMailHTML()
	body := msg
	msg1 := to + subject + mimeType + body
	return []byte(msg1)
}
