package domain

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func NewID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}
