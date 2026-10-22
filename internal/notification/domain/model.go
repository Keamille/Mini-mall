package domain

import (
	"encoding/json"
	"errors"
	"time"
)

type LogRow struct {
	ID        string     `db:"id" json:"id"`
	UserID    string     `db:"user_id" json:"user_id"`
	Channel   string     `db:"channel" json:"channel"`
	Template  string     `db:"template" json:"template"`
	Payload   []byte     `db:"payload" json:"-"`
	Status    int        `db:"status" json:"status"`
	EventID   string     `db:"event_id" json:"event_id"`
	ReadAt    *time.Time `db:"read_at" json:"read_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

type GinLog struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Channel   string          `json:"channel"`
	Template  string          `json:"template"`
	Payload   json.RawMessage `json:"payload"`
	Status    int             `json:"status"`
	EventID   string          `json:"event_id"`
	ReadAt    *time.Time      `json:"read_at"`
	CreatedAt time.Time       `json:"created_at"`
}

func (row LogRow) Public() GinLog {
	return GinLog{
		ID:        row.ID,
		UserID:    row.UserID,
		Channel:   row.Channel,
		Template:  row.Template,
		Payload:   json.RawMessage(row.Payload),
		Status:    row.Status,
		EventID:   row.EventID,
		ReadAt:    row.ReadAt,
		CreatedAt: row.CreatedAt,
	}
}

var (
	ErrTemplateNotFound = errors.New("template not found")
	ErrNotFound         = errors.New("notification not found")
)
