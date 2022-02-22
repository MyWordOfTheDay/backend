//go:build integration
// +build integration

package mail_test

import (
	"testing"

	"github.com/mywordoftheday/backend/internal/mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendMail(t *testing.T) {
	c, err := mail.New(mail.Config{
		SMTPHost:        "smtp.gmail.com",
		SMTPPort:        "587",
		SMTPUsername:    "simondrake1990@gmail.com",
		SMTPPassword:    "aevbeamcroqifoto",
		SMTPFromAddress: "simondrake1990@gmail.com",
		SMTPToAddresses: []string{"simondrake1990@gmail.com"},
	})

	require.NoError(t, err)

	assert.NoError(t, c.SendMail("this is a subject"))
}
