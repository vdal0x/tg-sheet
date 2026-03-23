package tg

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram/auth"
	tdtg "github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// AuthCallbacks provides user input during the Telegram auth flow.
// Each function is called at most once per Run if auth is needed.
// Return ("", err) to cancel auth and propagate the error.
type AuthCallbacks struct {
	Phone    func() (string, error)
	Code     func() (string, error)
	Password func() (string, error)
}

type callbackAuth struct {
	cbs AuthCallbacks
	log *zap.Logger
}

func (a callbackAuth) Phone(_ context.Context) (string, error) {
	phone, err := a.cbs.Phone()
	a.log.Info("auth: sending phone", zap.String("phone", phone), zap.Error(err))
	return phone, err
}

func (a callbackAuth) Code(_ context.Context, sentCode *tdtg.AuthSentCode) (string, error) {
	a.log.Info("auth: code requested",
		zap.String("type", fmt.Sprintf("%T", sentCode.Type)),
		zap.String("phone_code_hash", sentCode.PhoneCodeHash),
	)
	code, err := a.cbs.Code()
	a.log.Info("auth: code entered", zap.String("code", code), zap.Error(err))
	return code, err
}

func (a callbackAuth) Password(_ context.Context) (string, error) {
	a.log.Info("auth: 2FA password requested")
	pass, err := a.cbs.Password()
	a.log.Info("auth: password entered", zap.Error(err))
	return pass, err
}

func (a callbackAuth) AcceptTermsOfService(_ context.Context, tos tdtg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

func (a callbackAuth) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, &auth.SignUpRequired{}
}

func newAuthFlow(cbs AuthCallbacks, log *zap.Logger) auth.Flow {
	return auth.NewFlow(callbackAuth{cbs: cbs, log: log}, auth.SendCodeOptions{})
}
