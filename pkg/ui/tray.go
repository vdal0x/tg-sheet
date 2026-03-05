package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/getlantern/systray"
	"github.com/vdal0x/tg-sheet/pkg/config"
	"github.com/vdal0x/tg-sheet/pkg/parser"
	"github.com/vdal0x/tg-sheet/pkg/sheet"
	"github.com/vdal0x/tg-sheet/pkg/state"
	"github.com/vdal0x/tg-sheet/pkg/tg"
)

type TrayApp struct {
	cfg *config.Config
}

func NewTrayApp(cfg *config.Config) *TrayApp {
	return &TrayApp{cfg: cfg}
}

// Start blocks until the user quits. Must be called from the main goroutine.
func (a *TrayApp) Start() {
	systray.Run(a.onReady, func() {})
}

func (a *TrayApp) onReady() {
	systray.SetTitle("TGSheet")
	systray.SetTooltip("Telegram Sheet Reporter")

	mAuth := systray.AddMenuItem("Authenticate", "Connect to Telegram")
	mChats := systray.AddMenuItem("Select Chats", "Choose chats to track")
	mReport := systray.AddMenuItem("Generate Report", "Generate monthly CSV")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "")

	go func() {
		for {
			select {
			case <-mAuth.ClickedCh:
				go a.doAuth()
			case <-mChats.ClickedCh:
				go a.doSelectChats()
			case <-mReport.ClickedCh:
				go a.doGenerateReport()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

// newClient builds a TG client wired to osascript dialogs for auth input.
func (a *TrayApp) newClient() *tg.Client {
	return tg.NewClient(a.cfg.Tg.ApiId, a.cfg.Tg.ApiHash, tg.AuthCallbacks{
		Phone: func() string { return a.cfg.Tg.Phone },
		Code: func() string {
			code, _ := InputDialog("Telegram code:")
			return code
		},
		Password: func() string {
			pass, _ := InputDialog("2FA password:")
			return pass
		},
	})
}

func (a *TrayApp) doAuth() {
	client := a.newClient()
	err := client.Run(context.Background(), func(_ context.Context) error {
		return nil // auth is fully handled by IfNecessary before f is called
	})
	if err != nil {
		Notify("Auth failed", err.Error())
		return
	}
	Notify("TGSheet", "Authenticated successfully")
}

func (a *TrayApp) doSelectChats() {
	client := a.newClient()

	var chats []tg.Chat
	err := client.Run(context.Background(), func(ctx context.Context) error {
		var err error
		chats, err = client.FetchDialogs(ctx)
		return err
	})
	if err != nil {
		Notify("Error", err.Error())
		return
	}

	titles := make([]string, len(chats))
	for i, c := range chats {
		titles[i] = c.Title
	}

	indices, err := ChooseFromList("Select chats to track", titles)
	if err != nil || indices == nil {
		return // cancelled
	}

	st, err := state.Load()
	if err != nil {
		Notify("Error", err.Error())
		return
	}
	st.SelectedChatIDs = nil
	for _, i := range indices {
		st.SetSelected(chats[i].Id, true)
	}
	if err := st.Save(); err != nil {
		Notify("Error", err.Error())
		return
	}
	Notify("TGSheet", fmt.Sprintf("%d chat(s) selected", len(indices)))
}

func (a *TrayApp) doGenerateReport() {
	currentMonth := time.Now().Format("2006-01")
	monthStr, err := InputDialog(fmt.Sprintf("Month (YYYY-MM, default: %s):", currentMonth))
	if err != nil {
		return // cancelled
	}
	if monthStr == "" {
		monthStr = currentMonth
	}

	from, to, err := parseMonth(monthStr)
	if err != nil {
		Notify("Invalid month", "Expected format: YYYY-MM")
		return
	}

	st, err := state.Load()
	if err != nil {
		Notify("Error", err.Error())
		return
	}
	if len(st.SelectedChatIDs) == 0 {
		Notify("TGSheet", "No chats selected — use 'Select Chats' first")
		return
	}

	client := a.newClient()
	var allMsgs []tg.RawMessage

	err = client.Run(context.Background(), func(ctx context.Context) error {
		chats, err := client.FetchDialogs(ctx)
		if err != nil {
			return err
		}
		for _, c := range chats {
			if !st.IsSelected(c.Id) {
				continue
			}
			msgs, err := client.FetchMessages(ctx, c, from, to)
			if err != nil {
				return fmt.Errorf("fetch %q: %w", c.Title, err)
			}
			allMsgs = append(allMsgs, msgs...)
		}
		return nil
	})
	if err != nil {
		Notify("Error", err.Error())
		return
	}

	days := parser.Parse(allMsgs)

	home, _ := os.UserHomeDir()
	outFile := filepath.Join(home, "Desktop", fmt.Sprintf("report-%s.csv", monthStr))

	sh := sheet.NewSheet(days, outFile)
	if err := sh.Save(outFile); err != nil {
		Notify("Error", err.Error())
		return
	}
	Notify("TGSheet", fmt.Sprintf("Saved report-%s.csv to Desktop (%d days)", monthStr, len(days)))
}

func parseMonth(s string) (from, to time.Time, err error) {
	t, err := time.ParseInLocation("2006-01", s, time.Local)
	if err != nil {
		return
	}
	from = t
	to = t.AddDate(0, 1, 0).Add(-time.Second)
	return
}
