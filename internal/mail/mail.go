package mail

import (
	"bytes"
	"embed"
	"fmt"
	pkgtemplate "html/template"
	"net/smtp"

	"github.com/pkg/errors"
)

type Config struct {
	SMTPHost        string
	SMTPPort        string
	SMTPUsername    string
	SMTPPassword    string
	SMTPFromAddress string
	SMTPToAddresses []string
}

type Client struct {
	auth smtp.Auth
	host string
	port string
	from string
	to   []string

	template *pkgtemplate.Template
}

// New accepts Config and an optional template and returns a configered Client
//
// If a template is not required, simply pass an empty string
func New(c Config, template embed.FS, patterns ...string) (*Client, error) {
	auth := smtp.PlainAuth("", c.SMTPFromAddress, c.SMTPPassword, c.SMTPHost)

	t, err := pkgtemplate.ParseFS(template, patterns...)
	if err != nil {
		return nil, err
	}

	return &Client{
		auth: auth,
		host: c.SMTPHost,
		port: c.SMTPPort,
		from: c.SMTPFromAddress,
		to:   c.SMTPToAddresses,

		template: t,
	}, nil
}

func (c *Client) SendMailFromTemplate(subject string, data interface{}) error {
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	var body bytes.Buffer
	body.Write([]byte(fmt.Sprintf("Subject: %s \n%s\n\n", subject, mimeHeaders)))

	if err := c.template.Execute(&body, data); err != nil {
		return errors.Wrap(err, "error executing template")
	}

	addr := fmt.Sprintf("%s:%s", c.host, c.port)
	return smtp.SendMail(addr, c.auth, c.from, c.to, body.Bytes())
}
