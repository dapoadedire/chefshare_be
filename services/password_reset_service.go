package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/resend/resend-go/v2"
)

// SendPasswordResetEmail sends an email with the OTP for password reset
func (s *EmailService) SendPasswordResetEmail(email, name, otp string) (string, error) {
	ctx := context.Background()
	currentYear := time.Now().Year()
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "no-reply@chefshare.app"
	}

	replyTo := os.Getenv("EMAIL_REPLY_TO")
	if replyTo == "" {
		replyTo = "support@chefshare.app"
	}

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Reset Your Chefshare Password</title>
	<style>
		@media only screen and (max-width: 600px) {
			.container {
				width: 100%% !important;
				padding: 20px 10px !important;
			}
		}
		body {
			margin: 0;
			padding: 0;
			font-family: Arial, sans-serif;
			background-color: #f4f4f4;
		}
		.container {
			width: 80%%;
			max-width: 600px;
			margin: 0 auto;
			background: white;
			padding: 30px;
			border-radius: 8px;
			box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1);
		}
		.header {
			text-align: center;
			padding-bottom: 20px;
			border-bottom: 1px solid #e0e0e0;
		}
		.content {
			padding: 30px 0;
		}
		.otp-container {
			text-align: center;
			margin: 20px 0;
			padding: 15px;
			background-color: #f8f8f8;
			border-radius: 5px;
		}
		.otp-code {
			font-size: 32px;
			font-weight: bold;
			letter-spacing: 5px;
			color: #333;
		}
		.note {
			margin-top: 20px;
			padding: 15px;
			background-color: #fff8db;
			border-left: 4px solid #ffe066;
			color: #5c5c5c;
		}
		.footer {
			text-align: center;
			padding-top: 20px;
			border-top: 1px solid #e0e0e0;
			color: #7f8c8d;
			font-size: 12px;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>Password Reset</h2>
		</div>
		<div class="content">
			<p>Hi %s,</p>
			<p>We received a request to reset your Chefshare password. Use the following verification code to complete the process:</p>
			
			<div class="otp-container">
				<div class="otp-code">%s</div>
			</div>
			
			<p>This code is valid for 15 minutes and can only be used once.</p>
			
			<div class="note">
				<p>If you didn't request a password reset, please ignore this email or contact us if you have concerns.</p>
			</div>
		</div>
		<div class="footer">
			<p>This is an automated message, please do not reply directly.</p>
			<p>&copy; %d Chefshare. All rights reserved.</p>
		</div>
	</div>
</body>
</html>
`, name, otp, currentYear)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Chefshare <%s>", from),
		To:      []string{email},
		Subject: "Password Reset Code - Chefshare",
		Html:    htmlContent,
		ReplyTo: fmt.Sprintf("Chefshare <%s>", replyTo),
	}

	sent, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		log.Printf("Failed to send password reset email to %s: %v", email, err)
		return "", err
	}

	return sent.Id, nil
}

// SendPasswordChangedEmail notifies the user that their password has been changed
func (s *EmailService) SendPasswordChangedEmail(email, name string) (string, error) {
	ctx := context.Background()
	currentYear := time.Now().Year()
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "no-reply@chefshare.app"
	}

	replyTo := os.Getenv("EMAIL_REPLY_TO")
	if replyTo == "" {
		replyTo = "support@chefshare.app"
	}

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Your Chefshare Password Has Been Changed</title>
	<style>
		@media only screen and (max-width: 600px) {
			.container {
				width: 100%% !important;
				padding: 20px 10px !important;
			}
		}
		body {
			margin: 0;
			padding: 0;
			font-family: Arial, sans-serif;
			background-color: #f4f4f4;
		}
		.container {
			width: 80%%;
			max-width: 600px;
			margin: 0 auto;
			background: white;
			padding: 30px;
			border-radius: 8px;
			box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1);
		}
		.header {
			text-align: center;
			padding-bottom: 20px;
			border-bottom: 1px solid #e0e0e0;
		}
		.content {
			padding: 30px 0;
		}
		.alert {
			margin-top: 20px;
			padding: 15px;
			background-color: #e3f2fd;
			border-left: 4px solid #2196f3;
			color: #5c5c5c;
		}
		.footer {
			text-align: center;
			padding-top: 20px;
			border-top: 1px solid #e0e0e0;
			color: #7f8c8d;
			font-size: 12px;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>Password Changed Successfully</h2>
		</div>
		<div class="content">
			<p>Hi %s,</p>
			<p>This is a confirmation that your Chefshare account password has been successfully changed.</p>
			
			<div class="alert">
				<p>If you did not make this change, please contact our support team immediately.</p>
			</div>
			
			<p>You can now log in using your new password.</p>
		</div>
		<div class="footer">
			<p>This is an automated message, please do not reply directly.</p>
			<p>&copy; %d Chefshare. All rights reserved.</p>
		</div>
	</div>
</body>
</html>
`, name, currentYear)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Chefshare <%s>", from),
		To:      []string{email},
		Subject: "Your Password Has Been Changed - Chefshare",
		Html:    htmlContent,
		ReplyTo: fmt.Sprintf("Chefshare <%s>", replyTo),
	}

	sent, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		log.Printf("Failed to send password changed email to %s: %v", email, err)
		return "", err
	}

	return sent.Id, nil
}
