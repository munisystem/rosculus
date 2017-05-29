package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	sess *session.Session
)

func Session() *session.Session {
	if sess == nil {
		sess = session.Must(session.NewSession())
	}
	return sess
}
