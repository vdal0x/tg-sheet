package tg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	tdtg "github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type Client struct {
	appID   int
	appHash string
	cbs     AuthCallbacks
	log     *zap.Logger
	api     *tdtg.Client // set inside Run
}

func NewClient(appID int, appHash string, cbs AuthCallbacks, log *zap.Logger) *Client {
	return &Client{appID: appID, appHash: appHash, cbs: cbs, log: log}
}

// newTG creates a fresh telegram.Client backed by the persistent session file.
// The session directory is created on demand so FileStorage.StoreSession can write.
// gotd clients are one-shot — once Run() returns the instance is closed.
func (c *Client) newTG() *telegram.Client {
	path := sessionPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		c.log.Warn("could not create session dir", zap.Error(err))
	}
	return telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: path},
	})
}

// Authenticate runs the auth flow if needed and persists the session.
// If Telegram sends the code to the app (AuthSentCodeTypeApp), it
// automatically resends via SMS so the user gets a plain text code.
func (c *Client) Authenticate(ctx context.Context) error {
	tg := c.newTG()
	return tg.Run(ctx, func(ctx context.Context) error {
		return c.authFlow(ctx, tg.Auth())
	})
}

func (c *Client) authFlow(ctx context.Context, a *auth.Client) error {
	status, err := a.Status(ctx)
	if err != nil {
		return fmt.Errorf("check status: %w", err)
	}
	if status.Authorized {
		c.log.Info("already authenticated", zap.Int64("uid", status.User.ID))
		return nil
	}

	phone, err := c.cbs.Phone()
	if err != nil {
		return fmt.Errorf("get phone: %w", err)
	}

	sentClass, err := a.SendCode(ctx, phone, auth.SendCodeOptions{})
	if err != nil {
		return fmt.Errorf("send code: %w", err)
	}
	sent, ok := sentClass.(*tdtg.AuthSentCode)
	if !ok {
		return fmt.Errorf("unexpected sent code type: %T", sentClass)
	}
	c.log.Info("code sent", zap.String("type", fmt.Sprintf("%T", sent.Type)))

	// When the code goes to the Telegram app, resend it as SMS instead so the
	// user gets it without needing an active Telegram session.
	if _, isApp := sent.Type.(*tdtg.AuthSentCodeTypeApp); isApp {
		c.log.Info("code sent to app — requesting SMS resend")
		newClass, err := a.ResendCode(ctx, phone, sent.PhoneCodeHash)
		if err != nil {
			c.log.Warn("SMS resend failed, keeping app delivery", zap.Error(err))
		} else if newSent, ok := newClass.(*tdtg.AuthSentCode); ok {
			c.log.Info("resend ok", zap.String("type", fmt.Sprintf("%T", newSent.Type)))
			sent = newSent
		}
	}

	code, err := c.cbs.Code()
	if err != nil {
		return fmt.Errorf("get code: %w", err)
	}

	_, signInErr := a.SignIn(ctx, phone, code, sent.PhoneCodeHash)
	if errors.Is(signInErr, auth.ErrPasswordAuthNeeded) {
		pass, err := c.cbs.Password()
		if err != nil {
			return fmt.Errorf("get password: %w", err)
		}
		_, err = a.Password(ctx, pass)
		return err
	}
	return signInErr
}

// Run connects using the stored session and calls f.
// It does NOT perform authentication — if the session is missing or expired
// the API calls inside f will return an error (e.g. AUTH_KEY_UNREGISTERED).
// Call Authenticate first to create a valid session.
func (c *Client) Run(ctx context.Context, f func(ctx context.Context) error) error {
	tg := c.newTG()
	return tg.Run(ctx, func(ctx context.Context) error {
		c.api = tg.API()
		return f(ctx)
	})
}

func (c *Client) FetchDialogs(ctx context.Context) ([]Chat, error) {
	return FetchDialogs(ctx, c.api, c.log)
}

func (c *Client) FetchMessages(ctx context.Context, chat Chat, from, to time.Time) ([]RawMessage, error) {
	return FetchMessages(ctx, c.api, chat.Peer, from, to)
}


func sessionPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "tg-sheet", "session.json")
}
