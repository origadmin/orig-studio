package biz

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/mail.v2"
)

type EmailUseCase struct {
	settingUC *SettingUseCase
}

func NewEmailUseCase(settingUC *SettingUseCase) *EmailUseCase {
	return &EmailUseCase{settingUC: settingUC}
}

func (uc *EmailUseCase) SendEmail(ctx context.Context, to, subject, body string) error {
	host := uc.settingUC.Get(ctx, "smtp_host")
	if host == "" {
		return fmt.Errorf("SMTP not configured")
	}

	port := uc.settingUC.GetInt(ctx, "smtp_port")
	user := uc.settingUC.Get(ctx, "smtp_user")
	password := uc.settingUC.Get(ctx, "smtp_password")
	senderName := uc.settingUC.Get(ctx, "smtp_sender_name")
	if senderName == "" {
		senderName = "OrigStudio"
	}

	m := mail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", senderName, user))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(host, port, user, password)
	if port == 465 {
		d.SSL = true
	} else if uc.settingUC.Get(ctx, "smtp_use_tls") == "true" {
		d.StartTLSPolicy = mail.MandatoryStartTLS
	} else {
		d.StartTLSPolicy = mail.NoStartTLS
	}

	return d.DialAndSend(m)
}

func (uc *EmailUseCase) SendTestEmail(ctx context.Context, to string) error {
	siteName := uc.settingUC.Get(ctx, "site_name")
	if siteName == "" {
		siteName = "OrigStudio"
	}
	subject := siteName + " - Email Test"
	body := fmt.Sprintf(`<h2>Email Test</h2><p>If you received this email, SMTP is configured correctly.</p><p>From: %s</p>`, siteName)
	return uc.SendEmail(ctx, to, subject, body)
}

func (uc *EmailUseCase) IsConfigured(ctx context.Context) bool {
	return uc.settingUC.Get(ctx, "smtp_host") != "" &&
		uc.settingUC.Get(ctx, "smtp_user") != ""
}

func (uc *EmailUseCase) RenderTemplate(templateName string, data map[string]string) (string, string, error) {
	tmpl, ok := builtinTemplates[templateName]
	if !ok {
		return "", "", fmt.Errorf("template not found: %s", templateName)
	}

	subject := tmpl.Subject
	body := tmpl.BodyHTML
	for k, v := range data {
		subject = strings.ReplaceAll(subject, "{{."+k+"}}", v)
		body = strings.ReplaceAll(body, "{{."+k+"}}", v)
	}
	return subject, body, nil
}

// SMTPConfig contains the resolved SMTP configuration with defaults applied.
type SMTPConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	From        string `json:"from"`
	Nickname    string `json:"nickname"`
	SSL         bool   `json:"ssl"`
	Configured  bool   `json:"configured"`
	AutoDerived bool   `json:"auto_derived"`
}

// GetSMTPConfig returns the resolved SMTP configuration.
// If smtp_host is empty, derives defaults from the site domain.
func (uc *EmailUseCase) GetSMTPConfig(ctx context.Context) SMTPConfig {
	host := uc.settingUC.Get(ctx, "smtp_host")
	port := uc.settingUC.GetInt(ctx, "smtp_port")
	user := uc.settingUC.Get(ctx, "smtp_user")
	nickname := uc.settingUC.Get(ctx, "smtp_sender_name")
	if nickname == "" {
		nickname = "OrigStudio"
	}
	ssl := port == 465
	if !ssl && uc.settingUC.Get(ctx, "smtp_use_tls") == "true" {
		ssl = false
	}

	cfg := SMTPConfig{
		Host:     host,
		Port:     port,
		From:     user,
		Nickname: nickname,
		SSL:      ssl,
	}

	if host != "" && user != "" {
		cfg.Configured = true
		cfg.AutoDerived = false
		return cfg
	}

	domain := uc.settingUC.Get(ctx, "site_url")
	if domain == "" {
		domain = "example.com"
	}
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	if idx := strings.Index(domain, "/"); idx >= 0 {
		domain = domain[:idx]
	}

	cfg.Host = "smtp." + domain
	cfg.Port = 587
	cfg.From = "noreply@" + domain
	cfg.Nickname = nickname
	cfg.SSL = false
	cfg.Configured = false
	cfg.AutoDerived = true
	return cfg
}
