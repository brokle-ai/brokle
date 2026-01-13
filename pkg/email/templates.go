package email

import (
	"bytes"
	"fmt"
	"html/template"
)

// InvitationEmailParams contains parameters for invitation emails
type InvitationEmailParams struct {
	InviteeEmail     string // Email of the person being invited
	OrganizationName string // Name of the organization
	InviterName      string // Name of the person who sent the invite
	InviterEmail     string // Email of the person who sent the invite
	RoleName         string // Role being assigned
	PersonalMessage  string // Optional personal message from inviter
	AcceptURL        string // URL to accept the invitation
	ExpiresIn        string // Human-readable expiration (e.g., "7 days")
	AppName          string // Application name (e.g., "Brokle")
	AppURL           string // Application base URL
}

// BuildInvitationEmail generates the HTML email for a user invitation
func BuildInvitationEmail(params InvitationEmailParams) (html, text string, err error) {
	// HTML template
	htmlTmpl := template.Must(template.New("invitation_html").Parse(invitationHTMLTemplate))
	textTmpl := template.Must(template.New("invitation_text").Parse(invitationTextTemplate))

	// Set defaults
	if params.AppName == "" {
		params.AppName = "Brokle"
	}
	if params.ExpiresIn == "" {
		params.ExpiresIn = "7 days"
	}

	// Generate HTML
	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, params); err != nil {
		return "", "", fmt.Errorf("failed to generate HTML email: %w", err)
	}

	// Generate plain text
	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, params); err != nil {
		return "", "", fmt.Errorf("failed to generate text email: %w", err)
	}

	return htmlBuf.String(), textBuf.String(), nil
}

const invitationHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>You're invited to join {{.OrganizationName}}</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f4f4f5;">
  <table role="presentation" style="width: 100%; border-collapse: collapse;">
    <tr>
      <td align="center" style="padding: 40px 0;">
        <table role="presentation" style="width: 100%; max-width: 600px; border-collapse: collapse; background-color: #ffffff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
          <!-- Header -->
          <tr>
            <td style="padding: 40px 40px 20px 40px; text-align: center;">
              <h1 style="margin: 0; font-size: 24px; font-weight: 600; color: #18181b;">
                Join {{.OrganizationName}} on {{.AppName}}
              </h1>
            </td>
          </tr>

          <!-- Body -->
          <tr>
            <td style="padding: 20px 40px;">
              <p style="margin: 0 0 16px 0; font-size: 16px; line-height: 24px; color: #3f3f46;">
                Hi there,
              </p>
              <p style="margin: 0 0 16px 0; font-size: 16px; line-height: 24px; color: #3f3f46;">
                <strong>{{.InviterName}}</strong> ({{.InviterEmail}}) has invited you to join
                <strong>{{.OrganizationName}}</strong> as a <strong>{{.RoleName}}</strong>.
              </p>
              {{if .PersonalMessage}}
              <table role="presentation" style="width: 100%; border-collapse: collapse; margin: 20px 0;">
                <tr>
                  <td style="padding: 16px; background-color: #f4f4f5; border-radius: 6px; border-left: 4px solid #3b82f6;">
                    <p style="margin: 0; font-size: 14px; line-height: 22px; color: #52525b; font-style: italic;">
                      "{{.PersonalMessage}}"
                    </p>
                    <p style="margin: 8px 0 0 0; font-size: 13px; color: #71717a;">
                      — {{.InviterName}}
                    </p>
                  </td>
                </tr>
              </table>
              {{end}}
            </td>
          </tr>

          <!-- CTA Button -->
          <tr>
            <td style="padding: 20px 40px;">
              <table role="presentation" style="width: 100%; border-collapse: collapse;">
                <tr>
                  <td align="center">
                    <a href="{{.AcceptURL}}" style="display: inline-block; padding: 14px 32px; background-color: #18181b; color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 500; border-radius: 6px;">
                      Accept Invitation
                    </a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Expiration Notice -->
          <tr>
            <td style="padding: 0 40px 20px 40px;">
              <p style="margin: 0; font-size: 14px; line-height: 20px; color: #71717a; text-align: center;">
                This invitation expires in <strong>{{.ExpiresIn}}</strong>
              </p>
            </td>
          </tr>

          <!-- Divider -->
          <tr>
            <td style="padding: 0 40px;">
              <hr style="border: none; border-top: 1px solid #e4e4e7; margin: 0;">
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding: 20px 40px 40px 40px;">
              <p style="margin: 0 0 8px 0; font-size: 13px; line-height: 20px; color: #a1a1aa; text-align: center;">
                If you didn't expect this invitation, you can safely ignore this email.
              </p>
              <p style="margin: 0; font-size: 13px; line-height: 20px; color: #a1a1aa; text-align: center;">
                Button not working? Copy this link:<br>
                <a href="{{.AcceptURL}}" style="color: #3b82f6; word-break: break-all;">{{.AcceptURL}}</a>
              </p>
            </td>
          </tr>
        </table>

        <!-- Brand Footer -->
        <table role="presentation" style="width: 100%; max-width: 600px; border-collapse: collapse;">
          <tr>
            <td style="padding: 24px 40px; text-align: center;">
              <p style="margin: 0; font-size: 12px; color: #a1a1aa;">
                Sent by {{.AppName}} • AI Observability Platform
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

const invitationTextTemplate = `You're invited to join {{.OrganizationName}} on {{.AppName}}

Hi there,

{{.InviterName}} ({{.InviterEmail}}) has invited you to join {{.OrganizationName}} as a {{.RoleName}}.
{{if .PersonalMessage}}
Message from {{.InviterName}}:
"{{.PersonalMessage}}"
{{end}}
Accept your invitation by visiting:
{{.AcceptURL}}

This invitation expires in {{.ExpiresIn}}.

If you didn't expect this invitation, you can safely ignore this email.

---
Sent by {{.AppName}} - AI Observability Platform`
