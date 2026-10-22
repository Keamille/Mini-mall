package app

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/your-org/go-minimall/internal/user/domain"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (h *Handler) IssueRefreshToken(userID string) (string, error) {
	expiresAt := time.Now().UTC().Add(h.RefreshTTL)
	raw := domain.NewID() + domain.NewID()
	id := domain.NewID()
	return raw, h.Repo.InsertRefreshToken(id, userID, hashToken(raw), expiresAt)
}

func (h *Handler) respondWithTokens(c *gin.Context, status int, u domain.UserRow) {
	access, expiresIn, err := h.IssueAccessToken(u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "token issue failed"})
		return
	}
	refresh, err := h.IssueRefreshToken(u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "refresh issue failed"})
		return
	}
	c.JSON(status, gin.H{
		"user": gin.H{
			"id":       u.ID,
			"email":    u.Email.String,
			"phone":    u.Phone.String,
			"nickname": u.Nickname,
		},
		"access_token":  access,
		"refresh_token": refresh,
		"expires_in":    expiresIn,
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}

	row, err := h.Repo.FindRefreshTokenByHash(hashToken(req.RefreshToken))
	if err != nil || row.RevokedAt.Valid || time.Now().UTC().After(row.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 11003, "message": "invalid or expired refresh token"})
		return
	}

	_ = h.Repo.RevokeRefreshToken(row.ID)

	u, err := h.Repo.FindUserByID(row.UserID)
	if err != nil || u.Status != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 11003, "message": "invalid or expired refresh token"})
		return
	}

	h.respondWithTokens(c, http.StatusOK, u)
}
