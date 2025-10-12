package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// EmailService handles email sending
type EmailService struct {
	Config *EmailConfig
}

// NewEmailService creates a new email service
func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{Config: config}
}

// SendDiscountCode sends discount code email to user
func (s *EmailService) SendDiscountCode(email, discountCode string, discountPercent int, lang string) error {
	if !s.Config.Enabled {
		return nil // Email disabled, skip silently
	}

	// Select subject based on language
	subject := s.Config.SubjectEN
	if lang == "nl" {
		subject = s.Config.SubjectNL
	}

	// Generate HTML email content
	htmlContent, err := s.generateEmailHTML(discountCode, discountPercent, lang)
	if err != nil {
		return fmt.Errorf("failed to generate email HTML: %w", err)
	}

	// Send via Resend
	return s.sendViaResend(email, subject, htmlContent)
}

// sendViaResend sends email using Resend API
func (s *EmailService) sendViaResend(to, subject, htmlContent string) error {
	// Resend API endpoint
	url := "https://api.resend.com/emails"

	// Prepare request body
	payload := map[string]interface{}{
		"from":    fmt.Sprintf("%s <%s>", s.Config.FromName, s.Config.FromEmail),
		"to":      []string{to},
		"subject": subject,
		"html":    htmlContent,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email data: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+s.Config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("resend API error (status %d): %v", resp.StatusCode, errResp)
	}

	return nil
}

// generateEmailHTML creates beautiful HTML email
func (s *EmailService) generateEmailHTML(code string, percent int, lang string) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Subject}}</title>
</head>
<body style="margin: 0; padding: 0; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table width="100%" cellpadding="0" cellspacing="0" style="background-color: #f4f4f4; padding: 20px;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px 20px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px;">{{.Congrats}}</h1>
                        </td>
                    </tr>
                    
                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <p style="font-size: 16px; color: #333; margin: 0 0 20px;">{{.Message}}</p>
                            
                            <!-- Discount Code -->
                            <table width="100%" cellpadding="0" cellspacing="0" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); border-radius: 8px; margin: 30px 0;">
                                <tr>
                                    <td style="padding: 30px; text-align: center;">
                                        <p style="margin: 0 0 10px; color: #ffffff; font-size: 14px; text-transform: uppercase; letter-spacing: 1px;">{{.CodeLabel}}</p>
                                        <p style="margin: 0; color: #ffffff; font-size: 32px; font-weight: bold; letter-spacing: 3px;">{{.Code}}</p>
                                    </td>
                                </tr>
                            </table>
                            
                            <p style="font-size: 18px; color: #667eea; font-weight: 600; text-align: center; margin: 20px 0;">{{.SaveMessage}}</p>
                            
                            <p style="font-size: 14px; color: #666; margin: 30px 0 0; line-height: 1.6;">{{.Instructions}}</p>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f8f9fa; padding: 20px 30px; text-align: center; border-top: 1px solid #e0e0e0;">
                            <p style="margin: 0; font-size: 12px; color: #999;">{{.Footer}}</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`

	// Prepare template data based on language
	data := map[string]string{
		"Code": code,
	}

	if lang == "nl" {
		data["Subject"] = "Jouw Kortingscode"
		data["Congrats"] = "🎉 Gefeliciteerd!"
		data["Message"] = "Je hebt de puzzel voltooid! Hier is je kortingscode:"
		data["CodeLabel"] = "Jouw Kortingscode"
		data["SaveMessage"] = fmt.Sprintf("Bespaar %d%% op je volgende aankoop!", percent)
		data["Instructions"] = "Gebruik deze code bij het afrekenen om je korting te ontvangen. De code is geldig voor je volgende aankoop."
		data["Footer"] = "Bedankt voor het spelen van onze puzzel!"
	} else {
		data["Subject"] = "Your Discount Code"
		data["Congrats"] = "🎉 Congratulations!"
		data["Message"] = "You've completed the puzzle! Here's your discount code:"
		data["CodeLabel"] = "Your Discount Code"
		data["SaveMessage"] = fmt.Sprintf("Save %d%% on your next purchase!", percent)
		data["Instructions"] = "Use this code at checkout to receive your discount. The code is valid for your next purchase."
		data["Footer"] = "Thanks for playing our puzzle game!"
	}

	// Parse and execute template
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
