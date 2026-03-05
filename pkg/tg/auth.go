package tg

import (
	"context"

	"github.com/gotd/td/telegram/auth"
	tdtg "github.com/gotd/td/tg"
)

// AuthCallbacks provides user input during the Telegram auth flow.
// Each function is called at most once per Run if auth is needed.
type AuthCallbacks struct {
	Phone    func() string
	Code     func() string
	Password func() string // 2FA only
}

type callbackAuth struct {
	cbs AuthCallbacks
}

func (a callbackAuth) Phone(_ context.Context) (string, error) {
	return a.cbs.Phone(), nil
}

func (a callbackAuth) Code(_ context.Context, _ *tdtg.AuthSentCode) (string, error) {
	return a.cbs.Code(), nil
}

func (a callbackAuth) Password(_ context.Context) (string, error) {
	return a.cbs.Password(), nil
}

func (a callbackAuth) AcceptTermsOfService(_ context.Context, tos tdtg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

func (a callbackAuth) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, &auth.SignUpRequired{}
}

func newAuthFlow(cbs AuthCallbacks) auth.Flow {
	return auth.NewFlow(callbackAuth{cbs: cbs}, auth.SendCodeOptions{})
}
