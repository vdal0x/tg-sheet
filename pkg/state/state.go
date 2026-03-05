package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type State struct {
	SelectedChatIDs []int64 `json:"selected_chat_ids"`
}

func Load() (*State, error) {
	data, err := os.ReadFile(statePath())
	if os.IsNotExist(err) {
		return &State{}, nil
	}
	if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *State) Save() error {
	path := statePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (s *State) IsSelected(chatID int64) bool {
	for _, id := range s.SelectedChatIDs {
		if id == chatID {
			return true
		}
	}
	return false
}

func (s *State) SetSelected(chatID int64, selected bool) {
	if selected {
		if !s.IsSelected(chatID) {
			s.SelectedChatIDs = append(s.SelectedChatIDs, chatID)
		}
		return
	}
	ids := s.SelectedChatIDs[:0]
	for _, id := range s.SelectedChatIDs {
		if id != chatID {
			ids = append(ids, id)
		}
	}
	s.SelectedChatIDs = ids
}

func statePath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "tg-sheet", "state.json")
}
