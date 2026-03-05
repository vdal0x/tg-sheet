package tg

import (
	"context"
	"strings"
	"time"

	tdtg "github.com/gotd/td/tg"
)

type Chat struct {
	Id    int64
	Title string
	Peer  tdtg.InputPeerClass
}

type RawMessage struct {
	Id       int
	Date     time.Time
	Text     string
	FileName string // empty if not a file
}

// FetchDialogs returns up to 100 dialogs the user is a member of.
func FetchDialogs(ctx context.Context, tgCl *tdtg.Client) ([]Chat, error) {
	res, err := tgCl.MessagesGetDialogs(ctx, &tdtg.MessagesGetDialogsRequest{
		OffsetPeer: &tdtg.InputPeerEmpty{},
		Limit:      100,
	})
	if err != nil {
		return nil, err
	}

	var (
		chats      []tdtg.ChatClass
		users      []tdtg.UserClass
		rawDialogs []tdtg.DialogClass
	)

	switch v := res.(type) {
	case *tdtg.MessagesDialogs:
		chats, users, rawDialogs = v.Chats, v.Users, v.Dialogs
	case *tdtg.MessagesDialogsSlice:
		chats, users, rawDialogs = v.Chats, v.Users, v.Dialogs
	}

	chatByID := make(map[int64]tdtg.ChatClass, len(chats))
	for _, c := range chats {
		chatByID[c.GetID()] = c
	}
	userByID := make(map[int64]tdtg.UserClass, len(users))
	for _, u := range users {
		userByID[u.GetID()] = u
	}

	var dialogs []Chat
	for _, d := range rawDialogs {
		dlg, ok := d.(*tdtg.Dialog)
		if !ok {
			continue
		}
		if dialog, ok := buildDialog(dlg.Peer, chatByID, userByID); ok {
			dialogs = append(dialogs, dialog)
		}
	}
	return dialogs, nil
}

func buildDialog(peer tdtg.PeerClass, chatByID map[int64]tdtg.ChatClass, userByID map[int64]tdtg.UserClass) (Chat, bool) {
	switch p := peer.(type) {
	case *tdtg.PeerChannel:
		ch, ok := chatByID[p.ChannelID]
		if !ok {
			return Chat{}, false
		}
		c, ok := ch.(*tdtg.Channel)
		if !ok {
			return Chat{}, false
		}
		return Chat{
			Id:    c.ID,
			Title: c.Title,
			Peer:  &tdtg.InputPeerChannel{ChannelID: c.ID, AccessHash: c.AccessHash},
		}, true

	case *tdtg.PeerChat:
		ch, ok := chatByID[p.ChatID]
		if !ok {
			return Chat{}, false
		}
		c, ok := ch.(*tdtg.Chat)
		if !ok {
			return Chat{}, false
		}
		return Chat{
			Id:    c.ID,
			Title: c.Title,
			Peer:  &tdtg.InputPeerChat{ChatID: c.ID},
		}, true

	case *tdtg.PeerUser:
		u, ok := userByID[p.UserID]
		if !ok {
			return Chat{}, false
		}
		user, ok := u.(*tdtg.User)
		if !ok {
			return Chat{}, false
		}
		title := strings.TrimSpace(user.FirstName + " " + user.LastName)
		if title == "" {
			title = "@" + user.Username
		}
		return Chat{
			Id:    user.ID,
			Title: title,
			Peer:  &tdtg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash},
		}, true
	}
	return Chat{}, false
}

// FetchMessages returns all outgoing messages in [from, to] from the given peer.
func FetchMessages(ctx context.Context, tgCl *tdtg.Client, peer tdtg.InputPeerClass, from, to time.Time) ([]RawMessage, error) {
	var result []RawMessage

	req := &tdtg.MessagesGetHistoryRequest{
		Peer:       peer,
		OffsetDate: int(to.AddDate(0, 0, 1).Unix()),
		Limit:      100,
	}

	for {
		res, err := tgCl.MessagesGetHistory(ctx, req)
		if err != nil {
			return nil, err
		}

		msgs := extractMsgs(res)
		if len(msgs) == 0 {
			break
		}

		done := false
		for _, m := range msgs {
			msg, ok := m.(*tdtg.Message)
			if !ok {
				continue
			}
			t := time.Unix(int64(msg.Date), 0)
			if t.Before(from) {
				done = true
				break
			}
			if msg.Out {
				result = append(result, toRaw(msg, t))
			}
			req.OffsetID = msg.ID
		}

		req.OffsetDate = 0 // use Id for subsequent pages
		if done || len(msgs) < 100 {
			break
		}
	}

	return result, nil
}

func extractMsgs(res tdtg.MessagesMessagesClass) []tdtg.MessageClass {
	switch v := res.(type) {
	case *tdtg.MessagesMessages:
		return v.Messages
	case *tdtg.MessagesMessagesSlice:
		return v.Messages
	case *tdtg.MessagesChannelMessages:
		return v.Messages
	}
	return nil
}

func toRaw(msg *tdtg.Message, t time.Time) RawMessage {
	raw := RawMessage{Id: msg.ID, Date: t, Text: msg.Message}
	if media, ok := msg.Media.(*tdtg.MessageMediaDocument); ok {
		if doc, ok := media.Document.(*tdtg.Document); ok {
			for _, attr := range doc.Attributes {
				if fn, ok := attr.(*tdtg.DocumentAttributeFilename); ok {
					raw.FileName = fn.FileName
					break
				}
			}
		}
	}
	return raw
}
