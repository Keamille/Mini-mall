package port

import (
	"time"

	"github.com/your-org/go-minimall/internal/user/domain"
)

type Repository interface {
	ListAddresses(userID string) ([]domain.AddressRow, error)
	CreateAddressTx(a domain.AddressRow) error
	FindAddressByID(id string) (domain.AddressRow, error)
	UpdateAddressTx(a domain.AddressRow, setDefault bool) error
	DeleteAddress(id, userID string) error
	CreateUser(id string, email, phone any, passwordHash, nickname string) error
	FindUserByEmail(email string) (domain.UserRow, error)
	FindUserByPhone(phone string) (domain.UserRow, error)
	FindUserByID(id string) (domain.UserRow, error)
	InsertRefreshToken(id, userID, tokenHash string, expiresAt time.Time) error
	FindRefreshTokenByHash(tokenHash string) (domain.RefreshTokenRow, error)
	RevokeRefreshToken(id string) error
}
