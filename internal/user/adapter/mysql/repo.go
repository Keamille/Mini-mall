package mysqlrepo

import (
	"database/sql"
	"fmt"
	"github.com/your-org/go-minimall/internal/user/domain"
	"github.com/your-org/go-minimall/internal/user/port"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repo struct {
	DB *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{DB: db}
}

var _ port.Repository = (*Repo)(nil)

func (r *Repo) ListAddresses(userID string) ([]domain.AddressRow, error) {
	var rows []domain.AddressRow
	err := r.DB.Select(&rows, `
		SELECT id, user_id, receiver_name, phone, province, city, district, detail, is_default
		FROM user_addresses WHERE user_id = ?
		ORDER BY is_default DESC, created_at DESC`, userID)
	return rows, err
}

// createAddressTx：若 is_default=1，先清零同用户其它默认，再插入（同一事务）。
func (r *Repo) CreateAddressTx(a domain.AddressRow) error {
	return r.withTx(func(tx *sqlx.Tx) error {
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
func (r *Repo) FindAddressByID(id string) (domain.AddressRow, error) {
	var a domain.AddressRow
	err := r.DB.Get(&a, `
		SELECT id, user_id, receiver_name, phone, province, city, district, detail, is_default
		FROM user_addresses WHERE id = ?`, id)
	return a, err
}

// updateAddressTx：若设为默认，先清零同用户其它默认，再更新（同一事务）。
func (r *Repo) UpdateAddressTx(a domain.AddressRow, setDefault bool) error {
	return r.withTx(func(tx *sqlx.Tx) error {
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

func (r *Repo) withTx(fn func(tx *sqlx.Tx) error) error {
	tx, err := r.DB.Beginx()
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
func (r *Repo) DeleteAddress(id, userID string) error {
	res, err := r.DB.Exec(`DELETE FROM user_addresses WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows // 或自定义错误
	}
	return nil
}
func (r *Repo) CreateUser(id string, email, phone any, passwordHash, nickname string) error {
	_, err := r.DB.Exec(`
		INSERT INTO users (id, email, phone, password_hash, nickname, status)
		VALUES (?, ?, ?, ?, ?, 1)`,
		id, email, phone, passwordHash, nickname,
	)
	return err
}

func (r *Repo) FindUserByEmail(email string) (domain.UserRow, error) {
	var u domain.UserRow
	err := r.DB.Get(&u, `
		SELECT id, email, phone, password_hash, nickname, status
		FROM users WHERE email = ? LIMIT 1`, email)
	return u, err
}

func (r *Repo) FindUserByPhone(phone string) (domain.UserRow, error) {
	var u domain.UserRow
	err := r.DB.Get(&u, `
		SELECT id, email, phone, password_hash, nickname, status
		FROM users WHERE phone = ? LIMIT 1`, phone)
	return u, err
}

func (r *Repo) FindUserByID(id string) (domain.UserRow, error) {
	var u domain.UserRow
	err := r.DB.Get(&u, `
		SELECT id, email, phone, password_hash, nickname, status
		FROM users WHERE id = ?`, id)
	return u, err
}

func (r *Repo) InsertRefreshToken(id, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.DB.Exec(`
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES (?, ?, ?, ?)`,
		id, userID, tokenHash, expiresAt,
	)
	return err
}

func (r *Repo) FindRefreshTokenByHash(tokenHash string) (domain.RefreshTokenRow, error) {
	var row domain.RefreshTokenRow
	err := r.DB.Get(&row, `
		SELECT id, user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = ?
		LIMIT 1`, tokenHash)
	return row, err
}

func (r *Repo) RevokeRefreshToken(id string) error {
	_, err := r.DB.Exec(`
		UPDATE refresh_tokens SET revoked_at = UTC_TIMESTAMP(3) WHERE id = ?`, id)
	return err
}
