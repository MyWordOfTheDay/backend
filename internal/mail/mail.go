package mail

import (
	"bytes"
	"fmt"
	pkgtemplate "html/template"
	"net/smtp"
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
func New(c Config, template string) (*Client, error) {
	auth := smtp.PlainAuth("", c.SMTPFromAddress, c.SMTPPassword, c.SMTPHost)

	var t *pkgtemplate.Template
	if template != "" {
		var err error

		t, err = pkgtemplate.ParseFiles(template)
		if err != nil {
			return nil, err
		}
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

	c.template.Execute(&body, data)

	addr := fmt.Sprintf("%s:%s", c.host, c.port)
	return smtp.SendMail(addr, c.auth, c.from, c.to, body.Bytes())
}
