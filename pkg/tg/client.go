package tg

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	tdtg "github.com/gotd/td/tg"
)

type Client struct {
	tg  *telegram.Client
	cbs AuthCallbacks
	api *tdtg.Client // set inside Run
}

func NewClient(appID int, appHash string, cbs AuthCallbacks) *Client {
	tg := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: sessionPath()},
	})
	return &Client{tg: tg, cbs: cbs}
}

// Run authenticates if needed then calls f.
// FetchDialogs and FetchMessages are safe to call within f.
func (c *Client) Run(ctx context.Context, f func(ctx context.Context) error) error {
	return c.tg.Run(ctx, func(ctx context.Context) error {
		if err := c.tg.Auth().IfNecessary(ctx, newAuthFlow(c.cbs)); err != nil {
			return err
		}
		c.api = c.tg.API()
		return f(ctx)
	})
}

func (c *Client) FetchDialogs(ctx context.Context) ([]Chat, error) {
	return FetchDialogs(ctx, c.api)
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
