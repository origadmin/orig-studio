package biz

type EmailTemplate struct {
	Name     string
	Subject  string
	BodyHTML string
}

var builtinTemplates = map[string]EmailTemplate{
	"welcome": {
		Name:    "Welcome Email",
		Subject: "Welcome to {{.SiteName}}",
		BodyHTML: `<h2>Welcome to {{.SiteName}}!</h2>
<p>Hello {{.Username}},</p>
<p>Your account has been created successfully.</p>
<p><a href="{{.SiteURL}}">Get Started</a></p>`,
	},
	"email_verify": {
		Name:    "Email Verification",
		Subject: "{{.SiteName}} - Verify Your Email",
		BodyHTML: `<h2>Verify Your Email</h2>
<p>Please click the link below to verify your email:</p>
<p><a href="{{.VerifyURL}}">Verify Email</a></p>
<p>This link will expire in 24 hours.</p>`,
	},
	"password_reset": {
		Name:    "Password Reset",
		Subject: "{{.SiteName}} - Reset Your Password",
		BodyHTML: `<h2>Reset Your Password</h2>
<p>Please click the link below to reset your password:</p>
<p><a href="{{.ResetURL}}">Reset Password</a></p>
<p>This link will expire in 1 hour.</p>`,
	},
}
