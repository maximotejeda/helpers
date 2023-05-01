package auth

import "fmt"

// recoverEmailBody: format email to recover email
func RecoverEmailBody(token, email, code, host string) (body string) {
	body = fmt.Sprintf(`
	<html>
		<body>
			<h1>
				Recover Account
			</h1>
			<p>A new recovery process for the account was issued for the email %s</p>
			<p>to continue with the process click on the link <a href="https://%s/recover?token=%s">Recover My account</a> and introduce the next code</p>
			<h3>%s</h3>
			<p>If you did not ask for recovery please dont click any of the links here</p>
		</body>
	</html>
	`, email, host, token, code)
	return body
}

// activateEmailBody: format email body to activate account through email
func ActivateEmailBody(token, host string) (body string) {
	body = fmt.Sprintf(`
	<html>
		<body>
			<h1>
				Activate Account
			</h1>
			<p>Please activate your new account</p>
			<p>Click the link to activate your account <a href="http://%s/activate?token=%s">Activate </a></p>
		</body>
	</html>
	`, host, token)
	return body
}
