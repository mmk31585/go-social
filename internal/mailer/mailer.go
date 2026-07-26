package mailer

import "embed"

const (
	fromName            = "GoSocial"
	maxRetries          = 3
	UserWelcomeTemplate = "user_invitation"
)

var FS embed.FS

type Client interface {
	Send(templateFile string, username string, email string, data any, isSandbox bool) error
}
