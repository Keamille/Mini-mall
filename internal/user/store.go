package user

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type userRow struct {
	ID           string         `db:"id"`
	Email        sql.NullString `db:"email"`
	Phone        sql.NullString `db:"phone"`
	PasswordHash string         `db:"password_hash"`
	Nickname     string         `db:"nickname"`
	Status       int8           `db:"status"`
}

type refreshTokenRow struct {
	ID        string       `db:"id"`
	UserID    string       `db:"user_id"`
	ExpiresAt time.Time    `db:"expires_at"`
	RevokedAt sql.NullTime `db:"revoked_at"`
}

type addressRow struct {
	ID           string `db:"id"`
	UserID       string `db:"user_id"`
	ReceiverName string `db:"receiver_name"`
	Phone        string `db:"phone"`
	Province     string `db:"province"`
	City         string `db:"city"`
	District     string `db:"district"`
	Detail       string `db:"detail"`
	IsDefault    int8   `db:"is_default"`
}

func (h *Handler) listAddresses(userID string) ([]addressRow, error) {
	var rows []addressRow
	err := h.DB.Select(&rows, `
		SELECT id, user_id, receiver_name, phone, province, city, district, detail, is_default
		FROM user_addresses WHERE user_id = ?
		ORDER BY is_default DESC, created_at DESC`, userID)
	return rows, err
}
// createAddressTx：若 is_default=1，先清零同用户其它默认，再插入（同一事务）。
func (h *Handler) createAddressTx(a addressRow) error {
	return h.withTx(func(tx *sqlx.Tx) error {
		if a.IsDefault == 1 {
			if _, err := tx.Exec(`UPDATE user_addresses SET is_default=0 WHERE user_id=?`, a.UserID); err != nil {
				return err
			}
		}
		_, err := tx.Exec(`
			INSERT INTO user_addresses
			(id, user_id, receiver_name, phone, province, city, district, detail, is_default)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			a.ID, a.UserID, a.ReceiverName, a.Phone, a.Province, a.City, a.District, a.Detail, a.IsDefault,
		)
		return err
	})
}
func (h *Handler) findAddressByID(id string) (addressRow, error) {
	var a addressRow
	err := h.DB.Get(&a, `
		SELECT id, user_id, receiver_name, phone, province, city, district, detail, is_default
		FROM user_addresses WHERE id = ?`, id)
	return a, err
}
// updateAddressTx：若设为默认，先清零同用户其它默认，再更新（同一事务）。
func (h *Handler) updateAddressTx(a addressRow, setDefault bool) error {
	return h.withTx(func(tx *sqlx.Tx) error {
		if setDefault && a.IsDefault == 1 {
			if _, err := tx.Exec(`UPDATE user_addresses SET is_default=0 WHERE user_id=?`, a.UserID); err != nil {
				return err
			}
		}
		_, err := tx.Exec(`
			UPDATE user_addresses SET
				receiver_name=?, phone=?, province=?, city=?, district=?, detail=?, is_default=?
			WHERE id=? AND user_id=?`,
			a.ReceiverName, a.Phone, a.Province, a.City, a.District, a.Detail, a.IsDefault,
			a.ID, a.UserID,
		)
		return err
	})
}

func (h *Handler) withTx(fn func(tx *sqlx.Tx) error) error {
	tx, err := h.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
func (h *Handler) deleteAddress(id, userID string) error {
	res, err := h.DB.Exec(`DELETE FROM user_addresses WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows // 或自定义错误
	}
	return nil
}
func (h *Handler) createUser(id string, email, phone any, passwordHash, nickname string) error {
	_, err := h.DB.Exec(`
		INSERT INTO users (id, email, phone, password_hash, nickname, status)
		VALUES (?, ?, ?, ?, ?, 1)`,
		id, email, phone, passwordHash, nickname,
	)
	return err
}

func (h *Handler) findUserByEmail(email string) (userRow, error) {
	var u userRow
	err := h.DB.Get(&u, `
		SELECT id, email, phone, password_hash, nickname, status
		FROM users WHERE email = ? LIMIT 1`, email)
	return u, err
}

func (h *Handler) findUserByPhone(phone string) (userRow, error) {
	var u userRow
	err := h.DB.Get(&u, `
		SELECT id, email, phone, password_hash, nickname, status
		FROM users WHERE phone = ? LIMIT 1`, phone)
	return u, err
}

func (h *Handler) findUserByID(id string) (userRow, error) {
	var u userRow
	err := h.DB.Get(&u, `
		SELECT id, email, phone, password_hash, nickname, status
		FROM users WHERE id = ?`, id)
	return u, err
}

func (h *Handler) insertRefreshToken(id, userID, tokenHash string, expiresAt time.Time) error {
	_, err := h.DB.Exec(`
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES (?, ?, ?, ?)`,
		id, userID, tokenHash, expiresAt,
	)
	return err
}

func (h *Handler) findRefreshTokenByHash(tokenHash string) (refreshTokenRow, error) {
	var row refreshTokenRow
	err := h.DB.Get(&row, `
		SELECT id, user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = ?
		LIMIT 1`, tokenHash)
	return row, err
}

func (h *Handler) revokeRefreshToken(id string) error {
	_, err := h.DB.Exec(`
		UPDATE refresh_tokens SET revoked_at = UTC_TIMESTAMP(3) WHERE id = ?`, id)
	return err
}
