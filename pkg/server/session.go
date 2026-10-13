package server

import (
	"errors"
	"io"

	"github.com/emersion/go-smtp"
	"go.rtnl.ai/x/rlog"
)

type Session struct {
	From string
	To   []string
}

// AuthMechanisms returns supported authentication methods.
// For a public-facing inbound server receiving mail from external domains (Gmail, Yahoo),
// we generally leave this empty as external servers deliver mail anonymously.
func (s *Session) AuthMechanisms() []string {
	return []string{}
}

func (s *Session) Auth(mechanism, username, password string) (smtp.Session, error) {
	return nil, errors.New("authentication not supported")
}

// Mail handles the MAIL FROM command.
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	rlog.Info("MAIL FROM", "from", from)
	s.From = from
	return nil
}

// Rcpt handles the RCPT TO command.
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	rlog.Info("RCPT TO", "to", to)
	s.To = append(s.To, to)
	return nil
}

// Data handles the DATA command.
func (s *Session) Data(r io.Reader) (err error) {
	var buf []byte
	if buf, err = io.ReadAll(r); err != nil {
		return err
	}
	rlog.Info("DATA", "data", string(buf))
	return nil
}

// Reset resets the session state to initial values.
func (s *Session) Reset() {
	s.From = ""
	s.To = nil
}

// Logout is called when the client leaves the server.
func (s *Session) Logout() error {
	return nil
}
