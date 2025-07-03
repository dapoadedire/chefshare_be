package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client *resend.Client
}

func NewEmailService() (*EmailService, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("RESEND_API_KEY not set in environment")
	}

	client := resend.NewClient(apiKey)
	return &EmailService{
		client: client,
	}, nil
}

func (s *EmailService) SendWelcomeEmail(email string, name string) (string, error) {
	ctx := context.Background()
	currentYear := time.Now().Year()
	from := os.Getenv("EMAIL_FROM")
	replyTo := os.Getenv("EMAIL_REPLY_TO")

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Welcome to Chefshare</title>
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
		.cta {
			text-align: center;
			margin: 30px 0;
		}
		.cta a {
			display: inline-block;
			background-color: #27ae60;
			color: white;
			padding: 12px 24px;
			text-decoration: none;
			border-radius: 5px;
			font-weight: bold;
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
			<h2>Welcome to Chefshare</h2>
		</div>
		<div class="content">
			<p>Hi %s,</p>
			<p>Thanks for signing up. Chefshare is your space to create, manage, and explore recipes shared by cooks like you.</p>
			<p>You can start by uploading your first recipe or discovering what others are cooking.</p>
			<div class="cta">
				<a href="https://chefshare-2025.vercel.app/profile">Go to Profile</a>
			</div>
			<p>Need help or have feedback? Just reply to this email.</p>
			<p>Happy cooking!</p>
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
		Subject: "Welcome to Chefshare!",
		Html:    htmlContent,
		ReplyTo: fmt.Sprintf("Chefshare <%s>", replyTo),
		// ScheduledAt: "in 1 hour",
	}

	sent, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		log.Printf("Failed to send welcome email to %s: %v", email, err)
		return "", err
	}

	return sent.Id, nil
}

// SendVerificationEmail sends an email with a verification link to verify the user's email address
func (s *EmailService) SendVerificationEmail(email string, name string, token string) (string, error) {
	ctx := context.Background()
	currentYear := time.Now().Year()
	from := os.Getenv("EMAIL_FROM")
	replyTo := os.Getenv("EMAIL_REPLY_TO")

	// Get the frontend URL for verification from environment, default to localhost if not set
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	// Create verification URL with the token
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Verify Your Email Address</title>
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
		.cta {
			text-align: center;
			margin: 30px 0;
		}
		.cta a {
			display: inline-block;
			background-color: #27ae60;
			color: white;
			padding: 12px 24px;
			text-decoration: none;
			border-radius: 5px;
			font-weight: bold;
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
			<h2>Verify Your Email Address</h2>
		</div>
		<div class="content">
			<p>Hi %s,</p>
			<p>Thank you for registering with Chefshare. Please verify your email address to activate your account.</p>
			<p>This verification link will expire in 48 hours.</p>
			<div class="cta">
				<a href="%s">Verify Email Address</a>
			</div>
			<p>If you didn't create this account, you can safely ignore this email.</p>
			<p>Happy cooking!</p>
		</div>
		<div class="footer">
			<p>This is an automated message, please do not reply directly.</p>
			<p>&copy; %d Chefshare. All rights reserved.</p>
		</div>
	</div>
</body>
</html>
`, name, verificationURL, currentYear)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Chefshare <%s>", from),
		To:      []string{email},
		Subject: "Verify Your Email Address - Chefshare",
		Html:    htmlContent,
		ReplyTo: fmt.Sprintf("Chefshare <%s>", replyTo),
	}

	sent, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		log.Printf("Failed to send verification email to %s: %v", email, err)
		return "", err
	}

	return sent.Id, nil
}
