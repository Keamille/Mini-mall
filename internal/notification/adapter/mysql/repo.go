package mysqlrepo

import (
	"database/sql"
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/notification/domain"
	"github.com/your-org/go-minimall/internal/notification/port"

	"github.com/go-sql-driver/mysql"
)

type Repo struct {
	DB *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{DB: db}
}

var _ port.Repository = (*Repo)(nil)

func templateOf(eventType string) (string, error) {
	switch eventType {
	case "order_created", "order_paid", "order_cancelled", "order_closed":
		return eventType, nil
	default:
		return "", domain.ErrTemplateNotFound
	}
}

func (r *Repo) Record(eventID, eventType, userID string, payload []byte) (domain.GinLog, error) {
	template, err := templateOf(eventType)
	if err != nil {
		return domain.GinLog{}, err
	}
	if len(payload) == 0 {
		payload = []byte(`{}`)
	}
	row := domain.LogRow{
		ID:       domain.NewID(),
		UserID:   userID,
		Channel:  "inbox",
		Template: template,
		Payload:  payload,
		Status:   1,
		EventID:  eventID,
	}
	_, err = r.DB.Exec(`
		INSERT INTO notification_logs
		(id, user_id, channel, template, payload, status, event_id)
		VALUES (?, ?, 'inbox', ?, ?, 1, ?)`,
		row.ID, row.UserID, row.Template, row.Payload, row.EventID,
	)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return r.findByEventID(eventID)
		}
		return domain.GinLog{}, err
	}
	log.Printf("notify user=%s template=%s event_id=%s payload=%s", userID, template, eventID, payload)
	saved, err := r.findByEventID(eventID)
	if err != nil {
		return row.Public(), nil
	}
	return saved, nil
}

func (r *Repo) findByEventID(eventID string) (domain.GinLog, error) {
	var row domain.LogRow
	err := r.DB.Get(&row, `
		SELECT id, user_id, channel, template, payload, status, event_id, read_at, created_at
		FROM notification_logs WHERE event_id = ?`, eventID)
	if err != nil {
		return domain.GinLog{}, err
	}
	return row.Public(), nil
}

func (r *Repo) ListMine(userID string) ([]domain.GinLog, error) {
	var rows []domain.LogRow
	err := r.DB.Select(&rows, `
		SELECT id, user_id, channel, template, payload, status, event_id, read_at, created_at
		FROM notification_logs
		WHERE user_id = ? AND channel = 'inbox'
		ORDER BY created_at DESC
		LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.GinLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Public())
	}
	return out, nil
}

func (r *Repo) MarkRead(id, userID string) error {
	res, err := r.DB.Exec(`
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
	err = r.DB.Get(&exists, `SELECT 1 FROM notification_logs WHERE id = ? AND user_id = ?`, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}
