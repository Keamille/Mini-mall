package notification

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	errTemplateNotFound = errors.New("template not found")
	errNotFound         = errors.New("notification not found")
)

type logRow struct {
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

func (row logRow) public() ginLog {
	return ginLog{
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

type ginLog struct {
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

func templateOf(eventType string) (string, error) {
	switch eventType {
	case "order_created", "order_paid", "order_cancelled", "order_closed":
		return eventType, nil
	default:
		return "", errTemplateNotFound
	}
}

func (h *Handler) record(eventID, eventType, userID string, payload []byte) (ginLog, error) {
	template, err := templateOf(eventType)
	if err != nil {
		return ginLog{}, err
	}
	if len(payload) == 0 {
		payload = []byte(`{}`)
	}
	row := logRow{
		ID:       NewID(),
		UserID:   userID,
		Channel:  "inbox",
		Template: template,
		Payload:  payload,
		Status:   1,
		EventID:  eventID,
	}
	_, err = h.DB.Exec(`
		INSERT INTO notification_logs
		(id, user_id, channel, template, payload, status, event_id)
		VALUES (?, ?, 'inbox', ?, ?, 1, ?)`,
		row.ID, row.UserID, row.Template, row.Payload, row.EventID,
	)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return h.findByEventID(eventID)
		}
		return ginLog{}, err
	}
	log.Printf("notify user=%s template=%s event_id=%s payload=%s", userID, template, eventID, payload)
	saved, err := h.findByEventID(eventID)
	if err != nil {
		return row.public(), nil
	}
	return saved, nil
}

func (h *Handler) findByEventID(eventID string) (ginLog, error) {
	var row logRow
	err := h.DB.Get(&row, `
		SELECT id, user_id, channel, template, payload, status, event_id, read_at, created_at
		FROM notification_logs WHERE event_id = ?`, eventID)
	if err != nil {
		return ginLog{}, err
	}
	return row.public(), nil
}

func (h *Handler) listMine(userID string) ([]ginLog, error) {
	var rows []logRow
	err := h.DB.Select(&rows, `
		SELECT id, user_id, channel, template, payload, status, event_id, read_at, created_at
		FROM notification_logs
		WHERE user_id = ? AND channel = 'inbox'
		ORDER BY created_at DESC
		LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ginLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.public())
	}
	return out, nil
}

func (h *Handler) markRead(id, userID string) error {
	res, err := h.DB.Exec(`
		UPDATE notification_logs
		SET read_at = UTC_TIMESTAMP(3)
		WHERE id = ? AND user_id = ? AND read_at IS NULL`, id, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 1 {
		return nil
	}
	var exists int
	err = h.DB.Get(&exists, `SELECT 1 FROM notification_logs WHERE id = ? AND user_id = ?`, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return errNotFound
	}
	return err
}
